export type Category = {
  id: number;
  name: string;
  short_name: string;
  created_at: string;
  updated_at: string;
};

export type Subcategory = {
  id: number;
  category_id: number;
  name: string;
  created_at: string;
  updated_at: string;
};

export type Post = {
  id: number;
  category_id: number;
  subcategory_id: number;
  email: string;
  name: string;
  body: string;
  status: number;
  time_posted: number;
  time_posted_at: string;
  price: number;
  has_price: boolean;
  has_image: boolean;
  created_at: string;
  updated_at: string;
};

export type DataResponse<T> = {
  data: T;
};

export type PostsResponse = {
  data: Post[];
  meta: {
    total: number;
    limit: number;
    offset: number;
  };
};
