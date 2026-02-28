import Link from "next/link";
import { HomeFooter } from "@/components/home_footer";
import { HomeHeader } from "@/components/home_header";

export default function PostNotFound() {
	return (
		<table id="universe" className="supost-universe">
			<tbody>
				<HomeHeader />
				<tr>
					<td colSpan={4} id="hi-path">
						<div className="bread_crumb_header" id="bread_crumb_header">
							<Link href="/">SUpost</Link>
							{" » "}
							<Link href="/">Stanford, California</Link>
						</div>
					</td>
				</tr>
				<tr>
					<td colSpan={4} style={{ padding: 40, textAlign: "center" }}>
						<h2 style={{ color: "#284c9e", marginBottom: 16 }}>Post not found</h2>
						<p style={{ marginBottom: 20 }}>
							The post you&apos;re looking for doesn&apos;t exist or has been removed.
						</p>
						<Link href="/" style={{ color: "#0000cc" }}>
							← Back to SUpost
						</Link>
					</td>
				</tr>
				<HomeFooter />
			</tbody>
		</table>
	);
}
