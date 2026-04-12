import { NavItem } from "@/types/nav";
import Link from "next/link";
import {useState} from "react";


export function NavItemComponent({item,} : {item: NavItem}) {
    const isLogoutBtn = "Logout" === item.name;
    const hoverColor = isLogoutBtn ? "hover:text-red custom-h-underline-red" : "custom-h-underline";
    const className = "uppercase flex items-center " + hoverColor;
    const PName = item.name ? <span className="font-hairline pl-2 text-2xl">{item.name}</span> : null;
    const [isHover, setIsHover] = useState(false);

    if (item.href !== undefined) {
        return (<Link className={className} href={item.href}>
                {<item.icon />}
                {PName}
            </Link>);
    }

    if (item.hover !== undefined)
        return item.hover(<item.icon />);

    return (<button
                className={className}
                onClick={item.action}
                onMouseEnter={() => (setIsHover(true))}
                onMouseLeave={() => (setIsHover(false))}>
        <item.icon color={isHover && isLogoutBtn ? "red" : "black"}/>
        {PName}
    </button>);
}
