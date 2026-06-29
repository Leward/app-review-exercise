import { useEffect, useState } from 'react'

type Review = {
  SourceID: string
  Title: string
  Author: string
  Content: string
  Rating: number
  Date: string
}

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
  const [reviews, setReviews] = useState<Review[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let isMounted = true

    const loadReviews = async () => {
      try {
        const response = await fetch('/reviews')
        if (!response.ok) {
          throw new Error(`Request failed with status ${response.status}`)
        }

        const payload: unknown = await response.json()
        if (!Array.isArray(payload)) {
          throw new Error('Unexpected response payload')
        }

        if (!isMounted) {
          return
        }
        setReviews(payload as Review[])
      } catch (caughtError) {
        if (!isMounted) {
          return
        }

        setError(
          caughtError instanceof Error
            ? caughtError.message
            : 'Failed to load reviews',
        )
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    void loadReviews()

    return () => {
      isMounted = false
    }
  }, [])

  return (
    <main className="min-h-screen bg-slate-100 px-4 py-10 text-slate-900">
      <div className="mx-auto w-full max-w-4xl space-y-6">
        <header className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">
            Reviews
          </h1>
          <p className="mt-2 text-sm text-slate-600">
            Live list from <code>localhost:8080/reviews</code>
          </p>
        </header>

        {isLoading && (
          <div className="rounded-xl border border-slate-200 bg-white p-6 text-slate-600 shadow-sm">
            Loading reviews…
          </div>
        )}

        {!isLoading && error && (
          <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-red-700 shadow-sm">
            Could not load reviews: {error}
          </div>
        )}

        {!isLoading && !error && reviews.length === 0 && (
          <div className="rounded-xl border border-slate-200 bg-white p-6 text-slate-600 shadow-sm">
            No reviews were returned by the API.
          </div>
        )}

        {!isLoading && !error && reviews.length > 0 && (
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
