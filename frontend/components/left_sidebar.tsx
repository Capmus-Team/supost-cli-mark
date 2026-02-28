import { CategorySection } from "@/components/category_section";
import type { Category, Subcategory } from "@/types/marketplace";

type LeftSidebarProps = {
  categories: Category[];
  subcategoriesByCategory: Record<number, Subcategory[]>;
  lastUpdatedByCategory?: Record<number, string>;
};

export function LeftSidebar({
  categories,
  subcategoriesByCategory,
  lastUpdatedByCategory = {},
}: LeftSidebarProps) {
  return (
    <aside className="supost-left-nav">
      <div className="supost-post-to-div" id="postToDiv">
        <a href="/add">post to classifieds</a>
        <div className="at_stanford_required">@stanford.edu required</div>
      </div>

      <div className="supost-buy-job-post buy_job_post">
        <a href="http://jobs.supost.com/">post a job</a>
        <a href="http://housing.supost.com/">post housing</a>
        <a href="http://cars.supost.com/">post a car</a>
        <div className="buy_job_post_note">open for all emails</div>
      </div>

      <div className="supost-classifieds" id="classifieds">
        <table id="classifiedsOverview" className="category">
          <tbody>
            <tr>
              <td>
                <div className="classifiedOverviewHeader">
                  <a href="/search">overview</a>
                </div>
                {categories.map((category) => (
                  <div key={category.id} className="classifiedsOverviewRow">
                    <div className="classifiedsOverviewTitle">
                      <a href={`/search/cat/${category.id}`}>
                        {category.short_name}
                      </a>
                    </div>
                    <div className="timeAgoShort">
                      {lastUpdatedByCategory[category.id] ?? "—"}
                    </div>
                  </div>
                ))}
              </td>
            </tr>
          </tbody>
        </table>

        {categories.map((category) => (
          <CategorySection
            key={category.id}
            category={category}
            subcategories={subcategoriesByCategory[category.id] ?? []}
            lastUpdated={lastUpdatedByCategory[category.id]}
          />
        ))}

        <div className="supost-buy-job-post-lower buy_job_post_lower">
          <a href="http://jobs.supost.com/">post a job</a>
        </div>
      </div>
    </aside>
  );
}
