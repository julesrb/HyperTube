"use client";

import {useModal} from "@/context/ModalContext";
import LanguageDropdown from "@/components/nav/LanguageDropdown";
import React, {useState} from "react";
import Link from "next/link";
import {
    ExitDoorIcon,
    HypertubeLogo,
    HypertubeSmallLogo,
    LanguageIcon,
    RegisterIcon,
    SearchIcon,
    UserIcon
} from "@/components/Icons";
import {useAuth} from "@/context/AuthContext";
import {useResponsiveSize} from "@/script/utils";

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
    const {user, logout} = useAuth();
    const size = useResponsiveSize();
    const iconLogo = size === "xs" ? HypertubeSmallLogo : HypertubeLogo;

    const navItems: NavItem[] = user !== null ? [{
        name: "", icon: iconLogo, href: "/",
    }, {
        name: "Search", icon: SearchIcon, href: "/movies",
    }, {
        name: "Account", icon: UserIcon, href: "/users",
    }, {
        name: "Logout", icon: ExitDoorIcon, action: logout,
    }, {
        name: "", icon: LanguageIcon, hover: LanguageDropdown,
    },] : [{
        name: "", icon: iconLogo, href: "/",
    }, {
        name: "Search", icon: SearchIcon, href: "/movies",
    }, {
        name: "Sign In", icon: UserIcon, action: () => openModal({type: "signin"}),
    }, {
        name: "Create Account", icon: RegisterIcon, action: () => openModal({type: "register"}),
    }, {
        name: "", icon: LanguageIcon, hover: LanguageDropdown,
    },];

    return (<nav className="flex justify-between px-8 sm:px-12 xl:px-16 py-8">
        {navItems.map((item, index) => (<NavItemComponent key={index} item={item}/>))}
    </nav>)
}

function NavItemComponent({item,} : {item: NavItem}) {
    const isLogoutBtn = "Logout" === item.name;
    const hoverColor = isLogoutBtn ? "hover:text-red custom-underline-red" : "custom-underline";
    const className = "uppercase flex items-center " + hoverColor;
    const PName = item.name ? <span className="font-hairline pl-2 text-2xl hidden xl:block">{item.name}</span> : null;
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
