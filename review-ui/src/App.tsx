import { useReviewsStream } from './hooks/useReviewsStream'

function formatReviewDate(dateValue: string) {
  const parsedDate = new Date(dateValue)
  if (Number.isNaN(parsedDate.getTime())) {
    return dateValue
  }

  return parsedDate.toLocaleDateString()
}

function getRatingStars(rating: number) {
  const safeRating = Math.max(0, Math.min(5, Math.floor(rating)))
  return `${'★'.repeat(safeRating)}${'☆'.repeat(5 - safeRating)}`
}

function App() {
  const { reviews, isPending, loadError, refreshWarning } = useReviewsStream()
  const statusMessage = loadError
    ? `Could not load reviews: ${loadError}`
    : refreshWarning
      ? `Showing latest cached reviews. Refresh failed: ${refreshWarning}`
      : isPending
        ? 'Loading reviews…'
        : null
  const statusClasses =
    loadError
      ? 'border-red-200 bg-red-50 text-red-700'
      : refreshWarning
        ? 'border-amber-200 bg-amber-50 text-amber-700'
        : 'border-slate-200 bg-white text-slate-600'

  return (
    <main className="min-h-screen bg-slate-100 px-4 py-10 text-slate-900">
      <div className="mx-auto w-full max-w-4xl space-y-6">
        <header className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">
            Reviews
          </h1>
          <p className="mt-2 text-sm text-slate-600">
            Live list from <code>localhost:8080/reviews-async</code>
          </p>
        </header>

        {statusMessage && (
          <div
            aria-live="polite"
            className={`min-h-16 rounded-xl border p-4 text-sm shadow-sm ${statusClasses} flex items-center`}
          >
            {statusMessage}
          </div>
        )}

        {!isPending && !loadError && reviews.length === 0 && (
          <div className="rounded-xl border border-slate-200 bg-white p-6 text-slate-600 shadow-sm">
            No reviews were returned by the API.
          </div>
        )}

        {reviews.length > 0 && (
          <ul className="space-y-4">
            {reviews.map((review) => (
              <li
                key={`${review.SourceID}-${review.Date}`}
                className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm"
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <h2 className="text-xl font-semibold text-slate-900">
                    {review.Title}
                  </h2>
                  <span className="rounded-md bg-amber-50 px-2 py-1 text-sm font-medium text-amber-700">
                    {getRatingStars(review.Rating)}
                  </span>
                </div>
                <p className="mt-2 text-sm text-slate-500">
                  By {review.Author} • {formatReviewDate(review.Date)}
                </p>
                <p className="mt-4 whitespace-pre-wrap text-slate-700">
                  {review.Content}
                </p>
              </li>
            ))}
          </ul>
        )}
      </div>
    </main>
  )
}

export default App
