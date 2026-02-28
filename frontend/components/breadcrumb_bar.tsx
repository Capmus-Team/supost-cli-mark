export function BreadcrumbBar() {
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
        </div>
        <div className="current-time" id="time_header">
          {timeStr} - Updated
        </div>
      </td>
    </tr>
  );
}
