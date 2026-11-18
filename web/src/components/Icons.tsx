import type { ReactElement, SVGProps } from 'react'

type IconProps = SVGProps<SVGSVGElement> & { size?: number }
type IconComponent = (p: IconProps) => ReactElement

function base({ size = 24, ...rest }: IconProps) {
  return {
    width: size,
    height: size,
    viewBox: '0 0 24 24',
    fill: 'none',
    stroke: 'currentColor',
    strokeWidth: 1.6,
    strokeLinecap: 'round' as const,
    strokeLinejoin: 'round' as const,
    ...rest,
  }
}

export const Icons: Record<string, IconComponent> = {
  cpu: (p) => (
    <svg {...base(p)}>
      <rect x="6" y="6" width="12" height="12" rx="2.5" />
      <rect x="10" y="10" width="4" height="4" rx="1" />
      <path d="M9 2v3M15 2v3M9 19v3M15 19v3M2 9h3M2 15h3M19 9h3M19 15h3" />
    </svg>
  ),
  dna: (p) => (
    <svg {...base(p)}>
      <path d="M7 3c0 5 10 5 10 10s-10 5-10 8" />
      <path d="M17 3c0 5-10 5-10 10s10 5 10 8" />
      <path d="M8.5 7h7M8.5 17h7M10 12h4" />
    </svg>
  ),
  gear: (p) => (
    <svg {...base(p)}>
      <circle cx="12" cy="12" r="3" />
      <path d="M12 2v3M12 19v3M4.2 4.2l2.1 2.1M17.7 17.7l2.1 2.1M2 12h3M19 12h3M4.2 19.8l2.1-2.1M17.7 6.3l2.1-2.1" />
    </svg>
  ),
  atom: (p) => (
    <svg {...base(p)}>
      <circle cx="12" cy="12" r="1.6" />
      <ellipse cx="12" cy="12" rx="9" ry="3.6" />
      <ellipse cx="12" cy="12" rx="9" ry="3.6" transform="rotate(60 12 12)" />
      <ellipse cx="12" cy="12" rx="9" ry="3.6" transform="rotate(120 12 12)" />
    </svg>
  ),
  flask: (p) => (
    <svg {...base(p)}>
      <path d="M9 3h6M10 3v5.5L5.5 17A2.5 2.5 0 0 0 7.8 21h8.4a2.5 2.5 0 0 0 2.3-4L14 8.5V3" />
      <path d="M7.5 15h9" />
    </svg>
  ),
  globe: (p) => (
    <svg {...base(p)}>
      <circle cx="12" cy="12" r="9" />
      <path d="M3 12h18M12 3c2.5 2.6 3.8 5.7 3.8 9S14.5 18.4 12 21c-2.5-2.6-3.8-5.7-3.8-9S9.5 5.6 12 3z" />
    </svg>
  ),
  chart: (p) => (
    <svg {...base(p)}>
      <path d="M3 21h18" />
      <rect x="5" y="11" width="3.5" height="7" rx="1" />
      <rect x="10.5" y="6" width="3.5" height="12" rx="1" />
      <rect x="16" y="9" width="3.5" height="9" rx="1" />
    </svg>
  ),
  brain: (p) => (
    <svg {...base(p)}>
      <path d="M9 4.5A3 3 0 0 0 6 7.5a3 3 0 0 0-1 5.6A3 3 0 0 0 8 18a3 3 0 0 0 4 .5V5.5A3 3 0 0 0 9 4.5z" />
      <path d="M15 4.5a3 3 0 0 1 3 3 3 3 0 0 1 1 5.6A3 3 0 0 1 16 18a3 3 0 0 1-4 .5" />
      <path d="M9 9h1.5M13.5 9H15M9 13.5h1.5M13.5 13.5H15" />
    </svg>
  ),
  shield: (p) => (
    <svg {...base(p)}>
      <path d="M12 3l7 3v5.5c0 4.4-2.9 7.9-7 9.5-4.1-1.6-7-5.1-7-9.5V6l7-3z" />
      <path d="M9.2 12l2 2 3.6-3.8" />
    </svg>
  ),
  leaf: (p) => (
    <svg {...base(p)}>
      <path d="M5 19C5 11 11 5 20 4c1 9-5 15-13 15H5z" />
      <path d="M5 19c2-4 5-7 9-9" />
    </svg>
  ),
  building: (p) => (
    <svg {...base(p)}>
      <path d="M4 21V7l7-4 7 4v14" />
      <path d="M4 21h16M9 21v-5h4v5M8 9h1.5M13 9h1.5M8 13h1.5M13 13h1.5" />
    </svg>
  ),
  math: (p) => (
    <svg {...base(p)}>
      <path d="M4 6h6L4 18h6" />
      <path d="M14 8l6 8M20 8l-6 8" />
    </svg>
  ),
  search: (p) => (
    <svg {...base(p)}>
      <circle cx="11" cy="11" r="7" />
      <path d="M20 20l-3.5-3.5" />
    </svg>
  ),
  quote: (p) => (
    <svg {...base(p)}>
      <path d="M9 6H5.5A2.5 2.5 0 0 0 3 8.5V12a2.5 2.5 0 0 0 2.5 2.5H7v.5a3 3 0 0 1-3 3" />
      <path d="M20 6h-3.5A2.5 2.5 0 0 0 14 8.5V12a2.5 2.5 0 0 0 2.5 2.5H18v.5a3 3 0 0 1-3 3" />
    </svg>
  ),
  graph: (p) => (
    <svg {...base(p)}>
      <circle cx="6" cy="6" r="2.2" />
      <circle cx="18" cy="8" r="2.2" />
      <circle cx="12" cy="18" r="2.2" />
      <path d="M8 6.8l8 1M7 8l4 8M16.7 9.9L13 16.2" />
    </svg>
  ),
  spark: (p) => (
    <svg {...base(p)}>
      <path d="M12 3l1.9 5.1L19 10l-5.1 1.9L12 17l-1.9-5.1L5 10l5.1-1.9L12 3z" />
      <path d="M18.5 16l.8 2.2 2.2.8-2.2.8-.8 2.2-.8-2.2-2.2-.8 2.2-.8.8-2.2z" />
    </svg>
  ),
  arrowRight: (p) => (
    <svg {...base(p)}>
      <path d="M4 12h15M13 6l6 6-6 6" />
    </svg>
  ),
  arrowLeft: (p) => (
    <svg {...base(p)}>
      <path d="M20 12H5M11 18L5 12l6-6" />
    </svg>
  ),
  download: (p) => (
    <svg {...base(p)}>
      <path d="M12 3v12M7 11l5 5 5-5M4 20h16" />
    </svg>
  ),
  eye: (p) => (
    <svg {...base(p)}>
      <path d="M2.5 12s3.5-6.5 9.5-6.5S21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12z" />
      <circle cx="12" cy="12" r="2.8" />
    </svg>
  ),
  heart: (p) => (
    <svg {...base(p)}>
      <path d="M12 20s-7-4.4-7-9.2A4 4 0 0 1 12 7.6 4 4 0 0 1 19 10.8C19 15.6 12 20 12 20z" />
    </svg>
  ),
  plus: (p) => (
    <svg {...base(p)}>
      <path d="M12 5v14M5 12h14" />
    </svg>
  ),
  check: (p) => (
    <svg {...base(p)}>
      <path d="M4 12.5l5 5L20 6.5" />
    </svg>
  ),
  alert: (p) => (
    <svg {...base(p)}>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7.5v6M12 16.5v.5" />
    </svg>
  ),
  clock: (p) => (
    <svg {...base(p)}>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5.3l3.4 2" />
    </svg>
  ),
  users: (p) => (
    <svg {...base(p)}>
      <circle cx="9" cy="8" r="3.2" />
      <path d="M3 20c0-3.3 2.7-5.5 6-5.5s6 2.2 6 5.5" />
      <path d="M16 5.6a3.2 3.2 0 0 1 0 6M18 20c0-2.4-.9-4.2-2.5-5.1" />
    </svg>
  ),
  file: (p) => (
    <svg {...base(p)}>
      <path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8l-5-5z" />
      <path d="M14 3v5h5M9 13h6M9 17h4" />
    </svg>
  ),
  sort: (p) => (
    <svg {...base(p)}>
      <path d="M7 4v16M4 17l3 3 3-3M17 20V4M14 7l3-3 3 3" />
    </svg>
  ),
}

export type IconName = keyof typeof Icons

export function Icon({ name, ...rest }: IconProps & { name: string }) {
  const Component = Icons[name] ?? Icons.spark
  return <Component {...rest} />
}
