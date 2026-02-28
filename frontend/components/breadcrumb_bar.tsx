export type BreadcrumbItem = {
  label: string;
  href?: string;
};

type BreadcrumbBarProps = {
  /** Optional extra breadcrumb items after Stanford (e.g. category, subcategory, post title) */
  items?: BreadcrumbItem[];
};

export function BreadcrumbBar({ items = [] }: BreadcrumbBarProps = {}) {
  const now = new Date();
  const timeStr = now.toLocaleString("en-US", {
    weekday: "short",
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });

  return (
    <tr>
      <td colSpan={4} id="hi-path">
        <div className="bread_crumb_header" id="bread_crumb_header">
          <a href="/">SUpost</a>
          {" » "}
          <a href="/">Stanford, California</a>
          {items.map((item, i) => (
            <span key={i}>
              {" » "}
              {item.href ? (
                <a href={item.href}>{item.label}</a>
              ) : (
                <span>{item.label}</span>
              )}
            </span>
          ))}
        </div>
        <div id="time_header" className="current-time">
          {timeStr} - Updated
        </div>
      </td>
    </tr>
  );
}
