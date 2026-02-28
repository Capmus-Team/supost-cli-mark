import type { Post } from "@/types/marketplace";

type RecentPostsProps = {
  posts: Post[];
};

function formatPrice(post: Post) {
  if (!post.has_price) return "";
  return ` - $${post.price}`;
}

export function RecentPosts({ posts }: RecentPostsProps) {
  return (
    <div id="recentPosts">
      <div className="moduleTitle">
        <a href="/search">recently posted</a>
      </div>
      {posts.map((post) => (
        <div key={post.id} className="one-result">
          <a href={`/post/${post.id}`} className="post-link">
            {post.name}
            {formatPrice(post)}
          </a>{" "}
          <span className="verified">@stanford.edu</span>{" "}
          {post.has_image && (
            <span className="photo-tag">
              <img
                alt="Camera"
                className="icon_photo"
                src="/camera.gif"
              />
            </span>
          )}{" "}
          <span className="days-ago">{post.time_posted_at ?? "—"}</span>
        </div>
      ))}
    </div>
  );
}
