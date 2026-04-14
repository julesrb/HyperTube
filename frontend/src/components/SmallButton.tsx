export default function SmallButton({children, onClick} : {children: string, onClick: () => void}) {
    return (<button className="text-sm text-gray hover:underline hover:underline-gray" onClick={onClick}>{children}</button>);
}
