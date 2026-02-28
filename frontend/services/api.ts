import { cache } from "react";
import type {
  Category,
  DataResponse,
  PostsResponse,
  Subcategory,
} from "@/types/marketplace";

// On Vercel, use same origin. Locally, API runs on :8080.
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ??
  (process.env.VERCEL_URL ? `https://${process.env.VERCEL_URL}` : "http://localhost:8080");

async function fetchJSON<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    next: { revalidate: 15 },
  });
  if (!response.ok) {
    throw new Error(`API request failed: ${response.status} ${path}`);
  }
  return (await response.json()) as T;
}

export const getCategories = cache(async (): Promise<Category[]> => {
  const response = await fetchJSON<DataResponse<Category[]>>("/api/categories");
  return response.data;
});

export const getSubcategoriesByCategory = cache(
  async (categoryID: number): Promise<Subcategory[]> => {
    const response = await fetchJSON<DataResponse<Subcategory[]>>(
      `/api/subcategories?category_id=${categoryID}`,
    );
    return response.data;
  },
);

export async function getRecentPosts(limit = 50) {
  return fetchJSON<PostsResponse>(`/api/posts?limit=${limit}`);
}

export async function getFeaturedJobPosts(limit = 3) {
  return fetchJSON<PostsResponse>(
    `/api/posts?category_id=2&limit=${limit}&status=1`,
  );
}
