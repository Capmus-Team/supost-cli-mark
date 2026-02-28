/**
 * Maps category short_name to supost.com CSS class for styling.
 */
export function getCategoryCssClass(shortName: string): string {
	const normalized = shortName.toLowerCase().replace(/\s+/g, "_");
	const map: Record<string, string> = {
		housing: "housing",
		for_sale: "forsale",
		forsale: "forsale",
		"for sale": "forsale",
		jobs: "off_campus_jobs",
		off_campus_jobs: "off_campus_jobs",
		"jobs off-campus": "off_campus_jobs",
		personals: "personals",
		campus_job: "campus_jobs",
		"campus job": "campus_jobs",
		campus_jobs: "campus_jobs",
		community: "community",
		services: "services",
		resumes: "resumes",
	};
	return map[normalized] ?? map[shortName.toLowerCase()] ?? "forsale";
}

/**
 * Maps category short_name to supost.com icon CSS class (for sidebar/overview).
 */
export function getCategoryIconClass(shortName: string): string {
	const normalized = shortName.toLowerCase().replace(/\s+/g, "_");
	const map: Record<string, string> = {
		housing: "housing-icon",
		for_sale: "for-sale-icon",
		forsale: "for-sale-icon",
		"for sale": "for-sale-icon",
		jobs: "off-campus-jobs-icon",
		off_campus_jobs: "off-campus-jobs-icon",
		"jobs off-campus": "off-campus-jobs-icon",
		personals: "personals-icon",
		campus_job: "campus-jobs-icon",
		"campus job": "campus-jobs-icon",
		campus_jobs: "campus-jobs-icon",
		community: "community-icon",
		services: "services-icon",
		resumes: "resumes-icon",
	};
	return map[normalized] ?? map[shortName.toLowerCase()] ?? "for-sale-icon";
}
