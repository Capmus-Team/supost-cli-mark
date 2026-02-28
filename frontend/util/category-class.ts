/**
 * Maps category short_name to supost.com CSS class for styling.
 */
export function getCategoryCssClass(shortName: string): string {
  const normalized = shortName.toLowerCase().replace(/\s+/g, "_");
  const map: Record<string, string> = {
    housing: "housing",
    "for_sale": "forsale",
    forsale: "forsale",
    "for sale": "forsale",
    jobs: "off_campus_jobs",
    "off_campus_jobs": "off_campus_jobs",
    "jobs off-campus": "off_campus_jobs",
    personals: "personals",
    "campus_job": "campus_jobs",
    "campus job": "campus_jobs",
    campus_jobs: "campus_jobs",
    community: "community",
    services: "services",
    resumes: "resumes",
  };
  return map[normalized] ?? map[shortName.toLowerCase()] ?? "forsale";
}
