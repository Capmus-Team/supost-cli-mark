import { notFound } from "next/navigation";
import { BreadcrumbBar } from "@/components/breadcrumb_bar";
import { MessagePosterForm } from "@/components/message_poster_form";
import { PostImage } from "@/components/post_image";
import { ReserveButton } from "@/components/reserve_button";
import { HomeFooter } from "@/components/home_footer";
import { HomeHeader } from "@/components/home_header";
import { getPost } from "@/services/api";
import { getCategories, getSubcategoriesByCategory } from "@/services/api";
import type { Post } from "@/types/marketplace";
import { getCategoryCssClass } from "@/util/category-class";

export const revalidate = 15;

const SUPOST_IMAGE_BASE = "https://supost-prod.s3.amazonaws.com/posts";

/** Returns AWS S3 image URL(s) for post. SUpost uses post_{id}a, post_{id}b, post_{id}c, post_{id}d (up to 4). */
function getPostImageUrls(postId: number, hasImage: boolean): string[] {
	if (!hasImage) return [];
	const suffixes = ["a", "b", "c", "d"];
	return suffixes.map((s) => `${SUPOST_IMAGE_BASE}/${postId}/post_${postId}${s}`);
}

function formatPostDate(post: Post): string {
	const raw = post.time_posted_at;
	if (typeof raw === "string" && raw) {
		try {
			const d = new Date(raw);
			if (!isNaN(d.getTime())) {
				return d.toLocaleString("en-US", {
					weekday: "short",
					year: "numeric",
					month: "short",
					day: "numeric",
					hour: "numeric",
					minute: "2-digit",
				});
			}
		} catch {
			// fall through
		}
	}
	if (post.time_posted && typeof post.time_posted === "number") {
		const d = new Date(post.time_posted * 1000);
		return d.toLocaleString("en-US", {
			weekday: "short",
			year: "numeric",
			month: "short",
			day: "numeric",
			hour: "numeric",
			minute: "2-digit",
		});
	}
	return "—";
}

function formatPrice(post: Post): string {
	if (!post.has_price || post.price == null) return "";
	return `$${Number(post.price).toFixed(0)}`;
}

function truncateBreadcrumbTitle(name: string, maxLen = 40): string {
	if (name.length <= maxLen) return name;
	return `${name.slice(0, maxLen - 3)}...`;
}

/** Renders body text with <p> and <br> like original SUpost */
function PostBody({ body }: { body: string }) {
	const paragraphs = body.split(/\n\n+/).filter(Boolean);
	return (
		<div className="post-text">
			{paragraphs.map((para, i) => (
				<p key={i}>
					{para.split("\n").map((line, j) => (
						<span key={j}>
							{j > 0 && <br />}
							{line}
						</span>
					))}
				</p>
			))}
		</div>
	);
}

type PageProps = {
	params: Promise<{ id: string }>;
};

export default async function PostDetailPage({ params }: PageProps) {
	const { id } = await params;
	const postId = parseInt(id, 10);
	if (isNaN(postId) || postId <= 0) {
		notFound();
	}

	const post = await getPost(postId);
	if (!post) {
		notFound();
	}

	const [categories, subcategories] = await Promise.all([
		getCategories(),
		getSubcategoriesByCategory(post.category_id),
	]);

	const category = categories.find((c) => c.id === post.category_id);
	const subcategory = subcategories.find((s) => s.id === post.subcategory_id);
	const categoryCssClass = category ? getCategoryCssClass(category.short_name) : "forsale";

	const breadcrumbItems = [
		...(category
			? [{ label: category.short_name || category.name, href: `/search/cat/${category.id}` }]
			: []),
		...(subcategory
			? [{ label: subcategory.name, href: `/search/sub/${subcategory.id}` }]
			: []),
		{ label: truncateBreadcrumbTitle(post.name) },
	];

	const priceStr = formatPrice(post);
	const titleWithPrice = priceStr ? `${post.name} - ${priceStr}` : post.name;
	const reservePrice = post.has_price && post.price ? (post.price / 2).toFixed(2) : null;
	const imageUrls = getPostImageUrls(post.id, post.has_image);

	return (
		<table id="universe" className="supost-universe">
			<tbody>
				<HomeHeader />
				<BreadcrumbBar items={breadcrumbItems} />
				<tr>
					<td colSpan={4}>
						<div id="item-realm" className={categoryCssClass}>
							<div className={`post-rectangle ${categoryCssClass}`} />

							<div id="item-body" className={categoryCssClass}>
								<div className={`item-contour ${categoryCssClass}`}>
									<h2 id="posttitle" className={categoryCssClass}>
										{titleWithPrice}
										<span className="verified"> @stanford.edu</span>
									</h2>

									<div id="item-headers">
										<div className="replyto1">
											Reply to: <span>Use the form at the right to send messages to this user.</span>
										</div>
										<div className="item-date">
											Date: <span>{formatPostDate(post)}</span>
										</div>
										{post.has_price && (
											<div className="item-price">
												Price: <span>{formatPrice(post)}</span>
											</div>
										)}
									</div>

									<div id="item-content">
										<table className="post-table">
											<tbody>
												<tr style={{ verticalAlign: "top" }}>
													<td className="postBodyBox">
														<div className="post-details">
															{reservePrice && (
																<ReserveButton
																	reservePrice={reservePrice}
																	postId={post.id}
																/>
															)}

															{imageUrls.length > 0 && (
																<div className="post-photos">
																	{imageUrls.map((url, idx) => (
																		<PostImage
																			key={idx}
																			alt={`Photo ${idx}`}
																			className="post-photo"
																			src={url}
																		/>
																	))}
																</div>
															)}

															{post.body && <PostBody body={post.body} />}

															<div className="no-spam milder">
																please do not message this poster about other commercial services
															</div>
														</div>
													</td>

													<td className="messagePosterBox">
														{reservePrice && (
															<div className="reservation-explanation-section">
																<a href="#" id="why-reserve">
																	Why Reserve?
																</a>
																<ul id="benefits-of-reservation">
																	<li>- Stake Your Claim: Delist the Post</li>
																	<li>- Signal You&apos;re a Serious Buyer</li>
																	<li>- Get the Seller&apos;s Direct Contact: Know Who You&apos;re Dealing With</li>
																	<li className="link">
																		<a href="http://reserve.supost.com/" target="_blank" rel="noopener noreferrer">
																			How does it work?
																		</a>
																	</li>
																</ul>
															</div>
														)}

														<div id="message-box">
															<table id="message-table">
																<tbody>
																	<tr>
																		<td>
																			<div className="email-title">Message Poster</div>
																			<MessagePosterForm postId={post.id} />
																		</td>
																	</tr>
																</tbody>
															</table>
														</div>
													</td>
												</tr>
											</tbody>
										</table>
									</div>
								</div>
							</div>
						</div>
					</td>
				</tr>
				<HomeFooter />
			</tbody>
		</table>
	);
}
