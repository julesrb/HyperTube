"use client";

import React from "react";
import {ExitDoorIcon, HypertubResponsiveLogo, LanguageIcon, RegisterIcon, SearchIcon, UserIcon} from "@/components/Icons";
import {usePathname} from "@/i18n/navigation";
import {useTranslations} from "next-intl";
import ProfilePicture from "@/components/features/user/ProfilePicture";
import useAuth from "@/contexts/AuthContext";
import useModal from "@/contexts/ModalContext";
import NavItem from "@/components/layout/nav/NavItem";
import SwitchLanguage from "@/components/layout/nav/SwitchLanguage";

export interface iNavItem {
    name: string
    icon: ({color, size}: {
        selected: boolean
        color?: string
        size?: number
    }) => React.JSX.Element
    href?: string
    action?: () => void
    hover?: (Icon: ({selected}: {selected: boolean}) => React.JSX.Element) => React.JSX.Element
}

export default function Navbar() {
    const {openModal} = useModal();
    const {user, logout} = useAuth();
    const pathname = usePathname()
    const t = useTranslations("nav");

    const navItems: iNavItem[] = user ? [{
        name: "", icon: HypertubResponsiveLogo, href: "/"}, {
        name: t("search"), icon: SearchIcon, href: "/movies"}, {
        name: t("account"), icon: () => <ProfilePicture user={user} />, href: "/users"}, {
        name: t("logout"), icon: ExitDoorIcon, action: logout}, {
        name: "", icon: LanguageIcon, hover: SwitchLanguage,
    },] : [{
        name: "", icon: HypertubResponsiveLogo, href: "/"}, {
        name: t("search"), icon: SearchIcon, href: "/movies"}, {
        name: t("signin"), icon: UserIcon, action: () => openModal({type: "signin"})}, {
        name: t("createAccount"), icon: RegisterIcon, action: () => openModal({type: "register"})}, {
        name: "", icon: LanguageIcon, hover: SwitchLanguage,
    },];

    return (<nav className="flex justify-between px-6 sm:px-10 xl:px-16 py-4 sm:py-10">
        {navItems.map((item, index) => (<NavItem key={index} item={item} selected={pathname === item.href} logoutBtn={t("logout")} />))}
    </nav>)
}
