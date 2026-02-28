import Link from "next/link";
import type { Post } from "@/types/marketplace";

type PhotoStripProps = {
  posts: Post[];
};

function truncateTitle(value: string, maxLen = 22) {
  if (value.length <= maxLen) return value;
  return `${value.slice(0, maxLen - 3)}...`;
}

export function PhotoStrip({ posts }: PhotoStripProps) {
  const withImages = posts.filter((post) => post.has_image);
  const withoutImages = posts.filter((p) => !p.has_image);
  const items = [...withImages, ...withoutImages].slice(0, 4);

  return (
    <div className="post_photos clearfix" id="tickerTd">
      {items.map((post) => (
        <div key={post.id} className="post_photo" title={post.name}>
          <Link href={`/post/${post.id}`}>
            <div className="photo-placeholder photo_image_tag">
              #{post.id}
            </div>
          </Link>
          <div className="post_photo_title">
            <Link href={`/post/${post.id}`}>
              {truncateTitle(post.name)}
            </Link>
          </div>
          <span className="days-ago">
            {post.time_posted_at ?? "—"}
          </span>
        </div>
      ))}
    </div>
  );
}
