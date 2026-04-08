export default function CrossIcon({ color = "black", size = 30 }) {
    const fullColor = `var(--color-${color})`;

    return (<svg width={size} height={size} viewBox="0 0 36 36" fill="none" xmlns="http://www.w3.org/2000/svg">
            <line x1="9.32258" y1="8.613" x2="27.2582" y2="26.5486" stroke={fullColor}/>
            <line x1="27.2584" y1="9.32279" x2="9.32279" y2="27.2584" stroke={fullColor}/>
        </svg>
    );
}
