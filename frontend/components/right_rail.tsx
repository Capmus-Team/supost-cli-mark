import type { Post } from "@/types/marketplace";

type RightRailProps = {
  featuredJobs: Post[];
};

export function RightRail({ featuredJobs }: RightRailProps) {
  return (
    <div id="thoughts">
      <div className="moduleTitle">
        <a href="/search/cat/2">featured job posts</a>
      </div>

      <div className="job-posts-container">
        {featuredJobs.map((post) => (
          <div key={post.id} className="one-result">
            <a href={`/post/${post.id}`} className="post-link">
              {post.name}
            </a>
          </div>
        ))}
      </div>

      <div className="moduleTitle">
        <a href="http://pro.supost.com/">Events for Stanford</a>
      </div>

      <div className="job-posts-container">
        <div className="events-placeholder">Event feed preview</div>
      </div>

      <div className="homepage-right-column-info">
        <a
          href="https://www.craigslist.org/about/help/safety/"
          target="_blank"
          rel="noopener noreferrer"
        >
          Safety: If someone sends you a check, do not send them any money back.
          Never.
        </a>
        <div className="notSponsored">
          SUpost is not sponsored by, endorsed by, or affiliated with Stanford
          University.
        </div>
      </div>
    </div>
  );
}
