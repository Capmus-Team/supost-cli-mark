"use client";

import { useState } from "react";

type MessagePosterFormProps = {
	postId: number;
};

export function MessagePosterForm({ postId }: MessagePosterFormProps) {
	const [message, setMessage] = useState("");
	const [email, setEmail] = useState("");
	const [status, setStatus] = useState<"idle" | "sending" | "success" | "error">("idle");
	const [errorMsg, setErrorMsg] = useState("");

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setStatus("sending");
		setErrorMsg("");

		const apiBase =
			typeof window !== "undefined"
				? process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || "http://localhost:8080"
				: "http://localhost:8080";
		const base = apiBase.replace(/\/$/, "");

		try {
			const res = await fetch(`${base}/api/posts/${postId}/messages`, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ message, email }),
			});

			const data = await res.json().catch(() => ({}));

			if (!res.ok) {
				const msg = data?.error?.message ?? `Request failed (${res.status})`;
				setErrorMsg(msg);
				setStatus("error");
				return;
			}

			setStatus("success");
			setMessage("");
			setEmail("");
		} catch (err) {
			setErrorMsg(err instanceof Error ? err.message : "Failed to send");
			setStatus("error");
		}
	}

	return (
		<form className="new_message" id="new_message" onSubmit={handleSubmit}>
			<div className="email-label">Message:</div>
			<div>
				<textarea
					cols={40}
					rows={20}
					maxLength={32000}
					name="message"
					value={message}
					onChange={(e) => setMessage(e.target.value)}
					required
					disabled={status === "sending"}
				/>
			</div>
			<div className="email-label">Your Email:</div>
			<div>
				<input
					type="email"
					size={70}
					maxLength={70}
					name="email"
					value={email}
					onChange={(e) => setEmail(e.target.value)}
					required
					disabled={status === "sending"}
				/>
			</div>
			{errorMsg && (
				<div className="formError" style={{ color: "red", padding: "5px 0" }}>
					{errorMsg}
				</div>
			)}
			{status === "success" && (
				<div style={{ color: "#21B573", padding: "5px 0" }}>Message sent!</div>
			)}
			<div className="messagePosterBtn">
				<input
					type="submit"
					value={status === "sending" ? "Sending..." : "Send!"}
					disabled={status === "sending"}
					data-disable-with="Sending..."
				/>
			</div>
		</form>
	);
}
