"use client";

export default function ErrorPage({ reset }: { reset: () => void }) {
  return (
    <main className="state-page">
      <div className="state-box">
        Failed to load homepage data.
        <button type="button" onClick={reset} className="retry-btn">
          Retry
        </button>
      </div>
    </main>
  );
}
