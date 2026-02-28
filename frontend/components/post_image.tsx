"use client";

import { useState } from "react";

/**
 * Renders post image from SUpost AWS S3. Falls back to placeholder on error.
 */
export function PostImage({
	src,
	alt,
	className,
}: {
	src: string;
	alt: string;
	className?: string;
}) {
	const [errored, setErrored] = useState(false);

	if (errored) {
		return (
			<div
				className={`${className ?? ""} photo-placeholder`}
				style={{
					width: 340,
					height: 255,
					display: "flex",
					alignItems: "center",
					justifyContent: "center",
					border: "1px solid #d7dff0",
					background: "#f7f9ff",
					color: "#666",
					fontSize: 12,
				}}
			>
				No image available
			</div>
		);
	}

	return (
		// eslint-disable-next-line @next/next/no-img-element
		<img
			alt={alt}
			src={src}
			className={className}
			onError={() => setErrored(true)}
		/>
	);
}
