"use client";

import { useCategoryColumns } from "@/hooks/use-category-columns";
import { getCategoryCssClass, getCategoryIconClass } from "@/util/category-class";
import type { Category, Subcategory } from "@/types/marketplace";

type CategorySectionProps = {
  category: Category;
  subcategories: Subcategory[];
  lastUpdated?: string;
};

export function CategorySection({
  category,
  subcategories,
  lastUpdated,
}: CategorySectionProps) {
  const cssClass = getCategoryCssClass(category.short_name);
  const timeAgo = lastUpdated ? `(about ${lastUpdated})` : "";
  const columnCount = subcategories.length > 7 ? 2 : 1;
  const columns = useCategoryColumns(subcategories, columnCount);

  return (
    <table className={`${cssClass} category`}>
      <tbody>
        <tr>
          <td className="category_header" colSpan={2}>
            <div className={`category-header-title ${getCategoryIconClass(category.short_name)}`}>
              <a href={`/search/cat/${category.id}`}>
                {category.short_name}
              </a>
            </div>
            {timeAgo && (
              <span className="category-header-timeAgo">{timeAgo}</span>
            )}
          </td>
        </tr>
        {columns.length > 0 && (
          <tr>
            {columns.map((column, idx) => (
              <td
                key={idx}
                className={`categoryColumn cat-col-${idx}`}
              >
                <ul>
                  {column.map((sub) => (
                    <li key={sub.id}>
                      <a href={`/search/sub/${sub.id}`}>{sub.name}</a>
                    </li>
                  ))}
                </ul>
              </td>
            ))}
          </tr>
        )}
      </tbody>
    </table>
  );
}
