import Head from "next/head"
import { useState, useEffect } from "react"

export default function Home(props) {
  const [response, setResponse] = useState(null)

  useEffect(() => {
    fetch(process.env.NEXT_PUBLIC_BACKEND_URL)
      .then((resp) => resp.json())
      .then(setResponse)
      .catch(console.error)
  }, [])

  return (
    <div className="container">
      <Head>
        <title>Example Frontend</title>
        <link rel="icon" href="/favicon.ico" />
      </Head>

      <main>
        <h1>Response: {JSON.stringify(response, null, "  ")}</h1>
      </main>
    </div>
  )
}

export async function getStaticProps() {
  console.log(process.env.MESSAGE || "oops")
  return { props: {} }
}
