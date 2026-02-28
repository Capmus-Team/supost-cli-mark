"use client";

import { useState } from "react";

type ReserveButtonProps = {
	reservePrice: string;
	postId: number;
};

/**
 * Reserve button + arrow that expand to show reserve form (matches original SUpost behavior).
 * Clicking button or arrow hides them and shows the reserve form.
 */
export function ReserveButton({ reservePrice, postId }: ReserveButtonProps) {
	const [expanded, setExpanded] = useState(false);

	if (expanded) {
		const formUrl = `https://calaxes.formstack.com/forms/reserve?idpost=${postId}`;
		return (
			<div id="post-formstack" className="post-formstack" style={{ display: "block" }}>
				<h2>Reserve this post for ${reservePrice}</h2>
				<p style={{ margin: "10px 0" }}>
					SUpost retains the reservation fee.
				</p>
				<iframe
					title="Reserve form"
					src={formUrl}
					className="post-formstack-iframe"
					style={{
						width: "100%",
						minHeight: 500,
						border: "1px solid #ddd",
						borderRadius: 4,
					}}
				/>
			</div>
		);
	}

	return (
		<div className="stripe-reserve-button">
			<button
				className="stripe-button-el expand-formstack"
				type="button"
				onClick={() => setExpanded(true)}
			>
				<span style={{ display: "block", minHeight: 30 }}>
					Reserve for ${reservePrice}
				</span>
			</button>
			<span
				className="arrow-left expand-formstack"
				role="button"
				tabIndex={0}
				onClick={() => setExpanded(true)}
				onKeyDown={(e) => e.key === "Enter" && setExpanded(true)}
				aria-label="Expand reserve form"
			>
				&nbsp;
			</span>
		</div>
	);
}
