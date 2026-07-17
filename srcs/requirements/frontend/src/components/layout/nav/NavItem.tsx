import {useState} from "react";
import Link from "next/link";
import {iNavItem} from "@/components/layout/nav/Navbar";

export default function NavItem({item, selected, logoutBtn}: {item: iNavItem, selected: boolean, logoutBtn: string}) {
    const isLogoutBtn = item.name === logoutBtn;
    const className = "uppercase flex items-center";
    const PName = item.name ? <span style={{transform: "translateY(-1px)"}} className={"custom-underline pl-1 xl:pl-2 text-lg xl:text-2xl hidden md:block text-nowrap " + (selected ? "font-base font-light" : "font-hairline") + (isLogoutBtn ? " hover:text-red" : "")}>{item.name}</span> : null;
    const [isHover, setIsHover] = useState(false);

    if (item.href !== undefined) {
        return (<Link className={className} href={item.href}>
            {<item.icon selected={selected}/>}
            {PName}
        </Link>);
    }

    if (item.hover !== undefined)
        return item.hover(item.icon);

    return (<button
        className={className}
        onClick={item.action}
        onMouseEnter={() => (setIsHover(true))}
        onMouseLeave={() => (setIsHover(false))}>
        <item.icon selected={isHover && isLogoutBtn ? true : selected} color={isHover && isLogoutBtn ? "red" : "black"}/>
        {PName}
    </button>);
}
