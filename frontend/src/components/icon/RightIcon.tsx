export default function RightIcon({ color = "black", size = 13 }) {
    const fullColor = `var(--color-${color})`;

    return (<svg width={size} height={size} viewBox="0 0 16 27" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M1.97437 24.8514L12.4744 13.3514L1.47437 1.35144" stroke={fullColor} stroke-width="4"/>
        </svg>
    );
}
