export default function TextButton({children, onClick, className} : {children: string, onClick: () => void, className?: string}) {
    return (<button
        className={"text-sm font-light text-gray text-nowrap " + (className ?? " custom-underline")}
        onClick={onClick}>
        {children}
    </button>);
}
