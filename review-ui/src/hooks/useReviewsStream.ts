import { useEffect, useState } from 'react'

export type Review = {
  SourceID: string
  Title: string
  Author: string
  Content: string
  Rating: number
  Date: string
}

type ReviewsPayload = {
  reviews: Review[]
}

type RefreshErrorPayload = {
  error: string
}

type UseReviewsStreamResult = {
  reviews: Review[]
  isPending: boolean
  loadError: string | null
  refreshWarning: string | null
}

function toDateUnixMillis(value: string) {
  const parsedDate = new Date(value)
  const time = parsedDate.getTime()
  return Number.isNaN(time) ? 0 : time
}

function normalizeReviews(reviews: Review[]) {
  const deduped = new Map<string, Review>()
  for (const review of reviews) {
    const existing = deduped.get(review.SourceID)
    if (!existing) {
      deduped.set(review.SourceID, review)
      continue
    }
    if (toDateUnixMillis(review.Date) > toDateUnixMillis(existing.Date)) {
      deduped.set(review.SourceID, review)
    }
  }

  return [...deduped.values()].sort(
    (left, right) => toDateUnixMillis(right.Date) - toDateUnixMillis(left.Date),
  )
}

function parseReviewsPayload(data: string) {
  const payload: unknown = JSON.parse(data)
  if (!payload || typeof payload !== 'object') {
    throw new Error('Unexpected response payload')
  }

  const reviews = (payload as ReviewsPayload).reviews
  if (!Array.isArray(reviews)) {
    throw new Error('Unexpected response payload')
  }
  return reviews
}

function parseRefreshErrorPayload(data: string) {
  const payload: unknown = JSON.parse(data)
  if (!payload || typeof payload !== 'object') {
    return 'Failed to refresh reviews'
  }
  const error = (payload as RefreshErrorPayload).error
  if (typeof error !== 'string' || error.length === 0) {
    return 'Failed to refresh reviews'
  }
  return error
}

export function useReviewsStream(): UseReviewsStreamResult {
  const [reviews, setReviews] = useState<Review[]>([])
  const [isStreamComplete, setIsStreamComplete] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [refreshWarning, setRefreshWarning] = useState<string | null>(null)

  useEffect(() => {
    let isMounted = true
    let dataCount = 0

    const eventSource = new EventSource('/reviews-async')

    const handleLoadFailure = (caughtError: unknown) => {
      if (!isMounted) return
      setLoadError(
        caughtError instanceof Error ? caughtError.message : 'Failed to load reviews',
      )
      setRefreshWarning(null)
      setIsStreamComplete(true)
      eventSource.close()
    }

    const handleData = (event: Event) => {
      try {
        const nextReviews = parseReviewsPayload((event as MessageEvent).data)
        dataCount++
        if (!isMounted) return
        setReviews((current) => normalizeReviews([...nextReviews, ...current]))
        setLoadError(null)
        setRefreshWarning(null)
        if (dataCount >= 2) {
          setIsStreamComplete(true)
          eventSource.close()
        }
      } catch (caughtError) {
        handleLoadFailure(caughtError)
      }
    }

    const handleRefreshError = (event: Event) => {
      if (!isMounted) return
      const parsedError = parseRefreshErrorPayload((event as MessageEvent).data)
      if (dataCount === 0) {
        setLoadError(parsedError)
        setRefreshWarning(null)
      } else {
        setRefreshWarning(parsedError)
      }
      setIsStreamComplete(true)
      eventSource.close()
    }

    const handleEventSourceError = () => {
      if (!isMounted) return
      if (dataCount === 0) {
        setLoadError('Failed to load reviews')
      }
      setIsStreamComplete(true)
      eventSource.close()
    }

    eventSource.addEventListener('data', handleData)
    eventSource.addEventListener('refresh_error', handleRefreshError)
    eventSource.onerror = handleEventSourceError

    return () => {
      isMounted = false
      eventSource.removeEventListener('data', handleData)
      eventSource.removeEventListener('refresh_error', handleRefreshError)
      eventSource.close()
    }
  }, [])

  return {
    reviews,
    isPending: !isStreamComplete,
    loadError,
    refreshWarning,
  }
}
