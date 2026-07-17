import { initializeApp, getApps, getApp } from "@firebase/app"
import { getAuth, GoogleAuthProvider, signInWithPopup } from "@firebase/auth"

const required = [
  ["NEXT_PUBLIC_FIREBASE_API_KEY", process.env.NEXT_PUBLIC_FIREBASE_API_KEY],
  ["NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN", process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN],
  ["NEXT_PUBLIC_FIREBASE_PROJECT_ID", process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID],
] as const
const missing = required.filter(([, v]) => !v).map(([k]) => k)
if (missing.length) {
  throw new Error(`Firebase not configured: set ${missing.join(", ")} in environment`)
}

const firebaseConfig = {
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID,
  storageBucket: process.env.NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID,
  appId: process.env.NEXT_PUBLIC_FIREBASE_APP_ID,
}

const app = getApps().length === 0 ? initializeApp(firebaseConfig) : getApp()
const auth = getAuth(app)
const googleProvider = new GoogleAuthProvider()

export { auth, googleProvider, signInWithPopup, GoogleAuthProvider }
