import { BreadcrumbBar } from "@/components/breadcrumb_bar";
import { HomeFooter } from "@/components/home_footer";
import { HomeHeader } from "@/components/home_header";
import { LeftSidebar } from "@/components/left_sidebar";
import { PhotoStrip } from "@/components/photo_strip";
import { RecentPosts } from "@/components/recent_posts";
import { RightRail } from "@/components/right_rail";
import {
  getCategories,
  getFeaturedJobPosts,
  getRecentPosts,
  getSubcategories,
} from "@/services/api";

export const revalidate = 15;

function formatTimeAgo(seconds: number): string {
  if (seconds < 60) return `${seconds} secs`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)} mins`;
  if (seconds < 86400) return `about ${Math.floor(seconds / 3600)} hours`;
  if (seconds < 172800) return "1 day";
  return `${Math.floor(seconds / 86400)} days`;
}

export default async function HomePage() {
  const [categories, allSubcategories, recentPostsResponse, featuredJobsResponse] = await Promise.all([
    getCategories(),
    getSubcategories(),
    getRecentPosts(50),
    getFeaturedJobPosts(3),
  ]);

  const subcategoriesByCategory = allSubcategories.reduce<Record<number, typeof allSubcategories>>((acc, row) => {
    if (!acc[row.category_id]) {
      acc[row.category_id] = [];
    }
    acc[row.category_id].push(row);
    return acc;
  }, {});

  const posts = recentPostsResponse.data;
  const now = Math.floor(Date.now() / 1000);
  const lastUpdatedByCategory: Record<number, string> = {};
  for (const cat of categories) {
    const catPosts = posts.filter((p) => p.category_id === cat.id);
    if (catPosts.length > 0) {
      const mostRecent = catPosts.reduce(
        (a, b) => (a.time_posted > b.time_posted ? a : b),
      );
      lastUpdatedByCategory[cat.id] = formatTimeAgo(
        now - mostRecent.time_posted,
      );
    }
  }

  const postsWithTimeAgo = posts.map((p) => ({
    ...p,
    time_posted_at: formatTimeAgo(now - p.time_posted),
  }));

  return (
    <table id="universe" className="supost-universe">
      <tbody>
        <HomeHeader />
        <BreadcrumbBar />
        <tr>
          <td colSpan={4}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <tbody>
                <tr id="mainContent">
                  <td id="leftNavBar" style={{ verticalAlign: "top" }}>
                    <LeftSidebar
                      categories={categories}
                      subcategoriesByCategory={subcategoriesByCategory}
                      lastUpdatedByCategory={lastUpdatedByCategory}
                    />
                  </td>
                  <td id="indexBody" style={{ verticalAlign: "top", paddingLeft: "5px" }}>
                    <table id="indexBodyContent" style={{ width: "100%" }}>
                      <tbody>
                        <tr>
                          <td colSpan={2} id="tickerTd">
                            <PhotoStrip posts={postsWithTimeAgo} />
                          </td>
                        </tr>
                        <tr>
                          <td id="recentPosts" style={{ verticalAlign: "top" }}>
                            <RecentPosts posts={postsWithTimeAgo} />
                          </td>
                          <td id="thoughts" style={{ verticalAlign: "top" }}>
                            <RightRail featuredJobs={featuredJobsResponse.data} />
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
        <HomeFooter />
      </tbody>
    </table>
  );
}
