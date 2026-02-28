import { cache } from "react";
import { headers } from "next/headers";
import type {
  Category,
  DataResponse,
  PostsResponse,
  Subcategory,
} from "@/types/marketplace";

function trimTrailingSlash(value: string): string {
  return value.endsWith("/") ? value.slice(0, -1) : value;
}

async function resolveAPIBaseURL(): Promise<string> {
  const explicit = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  if (explicit) {
    return trimTrailingSlash(explicit);
  }

  const requestHeaders = await headers();
  const host = requestHeaders.get("x-forwarded-host") ?? requestHeaders.get("host");
  if (host) {
    const proto = requestHeaders.get("x-forwarded-proto") ?? "https";
    return `${proto}://${host}`;
  }

  if (process.env.VERCEL_URL) {
    return `https://${process.env.VERCEL_URL}`;
  }
  return "http://localhost:8080";
}

async function fetchJSON<T>(path: string): Promise<T> {
  const apiBaseURL = await resolveAPIBaseURL();
  const response = await fetch(`${apiBaseURL}${path}`, {
    next: { revalidate: 15 },
  });
  if (!response.ok) {
    throw new Error(`API request failed: ${response.status} ${path}`);
  }
  return (await response.json()) as T;
}

async function fetchJSONSafe<T>(path: string, fallback: T): Promise<T> {
  try {
    return await fetchJSON<T>(path);
  } catch {
    return fallback;
  }
}

export const getCategories = cache(async (): Promise<Category[]> => {
  const response = await fetchJSONSafe<DataResponse<Category[]>>(
    "/api/categories",
    { data: [] },
  );
  return response.data;
});

export const getSubcategoriesByCategory = cache(
  async (categoryID: number): Promise<Subcategory[]> => {
    const response = await fetchJSONSafe<DataResponse<Subcategory[]>>(
      `/api/subcategories?category_id=${categoryID}`,
      { data: [] },
    );
    return response.data;
  },
);

export async function getRecentPosts(limit = 50) {
  return fetchJSONSafe<PostsResponse>(`/api/posts?limit=${limit}`, {
    data: [],
    meta: { total: 0, limit, offset: 0 },
  });
}

export async function getFeaturedJobPosts(limit = 3) {
  return fetchJSONSafe<PostsResponse>(
    `/api/posts?category_id=2&limit=${limit}&status=1`,
    { data: [], meta: { total: 0, limit, offset: 0 } },
  );
}
