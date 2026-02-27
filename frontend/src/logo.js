export const LOGO_SVG = `
<svg width="256" height="256" viewBox="0 0 256 256" fill="none" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="bgGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#8B5CF6"/>
      <stop offset="100%" stop-color="#6366F1"/>
    </linearGradient>
    <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
      <feGaussianBlur stdDeviation="6" result="blur" />
      <feOffset dx="0" dy="4" result="offsetBlur" />
      <feMerge><feMergeNode in="offsetBlur" /><feMergeNode in="SourceGraphic" /></feMerge>
    </filter>
  </defs>
  <rect width="256" height="256" rx="60" fill="url(#bgGrad)" filter="url(#glow)" />
  <path d="M70 110C70 104.477 74.4772 100 80 100H120L135 115H176C181.523 115 186 119.477 186 125V186C186 191.523 181.523 196 176 196H80C74.4772 196 70 191.523 70 186V110Z" fill="white" />
  <g>
    <path d="M165 45L205 85M205 45L165 85" stroke="white" stroke-width="14" stroke-linecap="round" stroke-linejoin="round" />
    <circle cx="165" cy="45" r="5" fill="white" />
    <circle cx="205" cy="85" r="5" fill="white" />
    <circle cx="205" cy="45" r="5" fill="white" />
    <circle cx="165" cy="85" r="5" fill="white" />
  </g>
</svg>
`;
