"use client";

import {useModal} from "@/context/ModalContext";
import LanguageDropdown from "@/components/nav/LanguageDropdown";
import React, {useState} from "react";
import {Link} from "@/i18n/navigation";
import {ExitDoorIcon, HypertubResponsiveLogo, LanguageIcon, RegisterIcon, SearchIcon, UserIcon} from "@/components/Icons";
import {useAuth} from "@/context/AuthContext";
import {usePathname} from "@/i18n/navigation";
import {useTranslations} from "next-intl";

type NavItem = {
    name: string
    icon: ({color, size}: {
        selected: boolean
        color?: string
        size?: number
    }) => React.JSX.Element
    href?: string
    action?: () => void
    hover?: (Icon: ({selected}: {selected: boolean}) => React.JSX.Element) => React.JSX.Element
};

export default function Navbar() {
    const {openModal} = useModal();
    const {user, logout} = useAuth();
    const pathname = usePathname()
    const t = useTranslations("nav");

    const navItems: NavItem[] = user !== null ? [{
        name: "", icon: HypertubResponsiveLogo, href: "/",}, {
        name: t("search"), icon: SearchIcon, href: "/movies",}, {
        name: t("account"), icon: UserIcon, href: "/users",}, {
        name: t("logout"), icon: ExitDoorIcon, action: logout,}, {
        name: "", icon: LanguageIcon, hover: LanguageDropdown,
    },] : [{
        name: "", icon: HypertubResponsiveLogo, href: "/",}, {
        name: t("search"), icon: SearchIcon, href: "/movies",}, {
        name: t("signIn"), icon: UserIcon, action: () => openModal({type: "signin"}),}, {
        name: t("createAccount"), icon: RegisterIcon, action: () => openModal({type: "register"}),}, {
        name: "", icon: LanguageIcon, hover: LanguageDropdown,
    },];

    return (<nav className="flex justify-between px-6 sm:px-10 xl:px-16 py-8">
        {navItems.map((item, index) => (<NavItemComponent key={index} item={item} selected={pathname === item.href} logoutBtn={t("logout")} />))}
    </nav>)
}

function NavItemComponent({item, selected, logoutBtn}: {item: NavItem, selected: boolean, logoutBtn: string}) {
    const isLogoutBtn = item.name === logoutBtn;
    const hoverColor = isLogoutBtn ? "hover:text-red custom-underline-red" : "custom-underline";
    const className = "uppercase flex items-center " + hoverColor;
    const PName = item.name ? <span style={{transform: "translateY(-1px)"}} className={"pl-1 xl:pl-2 text-xl xl:text-2xl hidden md:block text-nowrap " + (selected ? "font-base font-light" : "font-hairline")}>{item.name}</span> : null;
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
