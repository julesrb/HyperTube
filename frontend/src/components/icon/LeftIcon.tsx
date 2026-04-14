export default function LeftIcon({ color = "black", size = 13 }) {
    const fullColor = `var(--color-${color})`;

    return (<svg width={size} height={size} viewBox="0 0 16 27" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M13.2107 1.34854L2.71069 12.8485L13.7107 24.8485" stroke={fullColor} stroke-width="4"/>
        </svg>
    );
}
