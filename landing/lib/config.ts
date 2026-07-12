// Central place for public (NEXT_PUBLIC_*) config read from the environment.
// Nothing sensitive here — these ship to the browser by design. No hardcoded
// business URLs/emails in components; set them in .env.local.

/** Lemon Squeezy hosted checkout link for the Pro plan. Empty if unconfigured. */
export const CHECKOUT_URL = process.env.NEXT_PUBLIC_LEMONSQUEEZY_CHECKOUT_URL ?? ""

/** Sales / enterprise contact address. */
export const SALES_EMAIL = process.env.NEXT_PUBLIC_SALES_EMAIL ?? "sales@limiter.io"
