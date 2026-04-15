export default function SmallButton({children, onClick, className} : {children: string, onClick: () => void, className?: string}) {
    return (<button
        className={"text-sm text-gray hover:underline hover:underline-gray " + className}
        onClick={onClick}
    >{children}</button>);
}
