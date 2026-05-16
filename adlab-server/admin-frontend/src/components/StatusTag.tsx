interface Props { status: string; size?: 'sm' | 'md' }

export default function StatusTag({ status, size = 'md' }: Props) {
  const isActive = status === 'active'
  return (
    <span
      className={`status-dot ${isActive ? 'active' : 'inactive'}`}
      style={{
        fontSize: size === 'sm' ? 12 : 13,
        gap: size === 'sm' ? 6 : 7,
      }}
    >
      {isActive ? 'Active' : 'Inactive'}
    </span>
  )
}
