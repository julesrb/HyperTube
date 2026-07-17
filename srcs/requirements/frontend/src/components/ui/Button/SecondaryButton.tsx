export default function SecondaryButton({children, onClick, className, onContextMenu} : {children: string, onClick?: () => void, className?: string, onContextMenu?: () => void}) {
    return (<button
        className={"uppercase text-nowrap px-3 sm:px-5 h-8 sm:h-10 bg-white text-sm sm:text-base xl:text-lg " + (onClick ? "text-black " : "text-gray ") + className}
        disabled={!onClick} onClick={onClick} onContextMenu={onContextMenu} >
        {children}
    </button>);
}
