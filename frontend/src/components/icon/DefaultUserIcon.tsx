export default function DefaultUserIcon({ className = "", color = "white", bg = "gray", size = 180 }) {
    const fullColor = `var(--color-${color})`;
    const fullBG = `var(--color-${bg})`;

    return (<svg width={size} height={size} className={className} viewBox="0 0 302 302" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="151" cy="151" r="151" fill={fullBG}/>
            <path d="M150.505 145.828C198.329 145.83 237.297 187.267 234.895 235.565C230.549 265.478 66.1054 265.478 66.1054 235.565C63.7031 187.266 102.678 145.828 150.505 145.828Z" fill={fullColor}/>
            <path d="M150.505 43C176.429 43.0025 197.449 63.9277 197.449 89.7399C197.449 115.552 176.429 136.477 150.505 136.48C124.578 136.48 103.56 115.554 103.56 89.7399C103.56 63.9262 124.578 43 150.505 43Z" fill={fullColor}/>
        </svg>
    );
}
