import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "SUpost: Marketplace for Stanford Students",
  description:
    "SUpost.com homepage recreation powered by Go API + Supabase.",
  icons: {
    icon: "/legacy/SUPostLogo.gif",
  },
  alternates: {
    canonical: "/",
  },
  openGraph: {
    title: "SUpost: Marketplace for Stanford Students",
    description:
      "Live categories and posts from Supabase through the Go backend.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <head>
        <link rel="stylesheet" href="/legacy/style.css" />
        <link rel="stylesheet" href="/legacy/stripe.css" />
        <link rel="stylesheet" href="/legacy/supost-clone.css" />
      </head>
      <body>{children}</body>
    </html>
  );
}
