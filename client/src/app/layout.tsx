import '../components/styles/Root.css'
export const metadata = {
  title: 'SeekTune',
  description: 'Reverse Lookup',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
