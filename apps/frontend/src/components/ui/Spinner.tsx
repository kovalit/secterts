export function Spinner({ size = 18 }: { size?: number }) {
  return (
    <span
      className="inline-block animate-spin rounded-full border-2 border-border border-t-accent"
      style={{ width: size, height: size }}
      aria-label="Загрузка"
    />
  )
}

export function PageLoader() {
  return (
    <div className="grid place-items-center py-24">
      <Spinner size={28} />
    </div>
  )
}
