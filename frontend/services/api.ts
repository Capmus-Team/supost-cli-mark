import { cache } from "react";
import { headers } from "next/headers";
import type {
	Category,
	DataResponse,
	Post,
	PostsResponse,
	Subcategory,
} from "@/types/marketplace";

const TAXONOMY_REVALIDATE_SECONDS = 60 * 60;
const POSTS_REVALIDATE_SECONDS = 15;

function trimTrailingSlash(value: string): string {
	return value.endsWith("/") ? value.slice(0, -1) : value;
}

async function resolveAPIBaseURL(): Promise<string> {
	const explicit = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
	if (explicit) {
		return trimTrailingSlash(explicit);
	}

	if (process.env.NODE_ENV === "development") {
		return "http://localhost:8080";
	}

	const requestHeaders = await headers();
	const host =
		requestHeaders.get("x-forwarded-host") ?? requestHeaders.get("host");
	if (host) {
		const proto = requestHeaders.get("x-forwarded-proto") ?? "https";
		return `${proto}://${host}`;
	}

	if (process.env.VERCEL_URL) {
		return `https://${process.env.VERCEL_URL}`;
	}
	return "http://localhost:8080";
}

async function fetchJSON<T>(path: string, revalidateSeconds: number): Promise<T> {
	const apiBaseURL = await resolveAPIBaseURL();
	const response = await fetch(`${apiBaseURL}${path}`, {
		next: { revalidate: revalidateSeconds },
	});
	if (!response.ok) {
		throw new Error(`API request failed: ${response.status} ${path}`);
	}
	return (await response.json()) as T;
}

async function fetchJSONSafe<T>(
	path: string,
	fallback: T,
	revalidateSeconds: number,
): Promise<T> {
	try {
		return await fetchJSON<T>(path, revalidateSeconds);
	} catch {
		return fallback;
	}
}

export const getCategories = cache(async (): Promise<Category[]> => {
	const response = await fetchJSONSafe<DataResponse<Category[]>>(
		"/api/categories",
		{ data: [] },
		TAXONOMY_REVALIDATE_SECONDS,
	);
	return response.data;
});

export const getSubcategories = cache(async (): Promise<Subcategory[]> => {
	const response = await fetchJSONSafe<DataResponse<Subcategory[]>>(
		"/api/subcategories",
		{ data: [] },
		TAXONOMY_REVALIDATE_SECONDS,
	);
	if (response.data.length > 0) {
		return response.data;
	}

	// Compatibility fallback: some backends require category_id for subcategory queries.
	const categories = await getCategories();
	const grouped = await Promise.all(
		categories.map((category) => getSubcategoriesByCategory(category.id)),
	);
	return grouped.flat();
});

export const getSubcategoriesByCategory = cache(
	async (categoryID: number): Promise<Subcategory[]> => {
		const response = await fetchJSONSafe<DataResponse<Subcategory[]>>(
			`/api/subcategories?category_id=${categoryID}`,
			{ data: [] },
			TAXONOMY_REVALIDATE_SECONDS,
		);
		return response.data;
	},
);

export async function getRecentPosts(limit = 50) {
	return fetchJSONSafe<PostsResponse>(
		`/api/posts?limit=${limit}`,
		{
			data: [],
			meta: { total: 0, limit, offset: 0 },
		},
		POSTS_REVALIDATE_SECONDS,
	);
}

export async function getFeaturedJobPosts(limit = 3) {
	return fetchJSONSafe<PostsResponse>(
		`/api/posts?category_id=2&limit=${limit}&status=1`,
		{ data: [], meta: { total: 0, limit, offset: 0 } },
		POSTS_REVALIDATE_SECONDS,
	);
}

export type PostResponse = {
	data: Post;
};

export async function getPost(id: number): Promise<Post | null> {
	try {
		const response = await fetchJSON<PostResponse>(
			`/api/posts/${id}`,
			POSTS_REVALIDATE_SECONDS,
		);
		return response.data;
	} catch {
		return null;
	}
}
