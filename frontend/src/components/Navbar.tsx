"use client";

import {useModal} from "@/context/ModalContext";
import LanguageDropdown from "@/components/LanguageDropdown";
import React, {useState} from "react";
import Link from "next/link";
import {ExitDoorIcon, HypertubeLogo, LanguageIcon, RegisterIcon, SearchIcon, UserIcon} from "@/components/Icon";

type NavItem = {
    name: string
    icon: ({color, size}: {
        color?: string
        size?: number
    }) => React.JSX.Element
    href?: string
    action?: () => void
    hover?: (Icon: React.JSX.Element) => React.JSX.Element
};

export default function Navbar() {
    const {openModal} = useModal();
    const [isLogin, setIsLogin] = useState(true);

    const navItems: NavItem[] = isLogin ? [{
        name: "", icon: HypertubeLogo, href: "/",
    }, {
        name: "Search", icon: SearchIcon, href: "/movies",
    }, {
        name: "Account", icon: UserIcon, href: "/users",
    }, {
        name: "Logout", icon: ExitDoorIcon, action: () => {
            setIsLogin(false);
        },
    }, {
        name: "", icon: LanguageIcon, hover: LanguageDropdown,
    },] : [{
        name: "", icon: HypertubeLogo, href: "/",
    }, {
        name: "Search", icon: SearchIcon, href: "/movies",
    }, {
        name: "Sign In", icon: UserIcon, action: () => openModal("signin"),
    }, {
        name: "Create Account", icon: RegisterIcon, action: () => openModal("register"),
    }, {
        name: "", icon: LanguageIcon, hover: LanguageDropdown,
    },];

    return (<nav className="flex justify-between px-16 py-8">
        {navItems.map((item, index) => (<NavItemComponent key={index} item={item}/>))}
    </nav>)
}



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
