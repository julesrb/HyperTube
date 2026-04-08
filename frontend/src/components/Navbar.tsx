"use client";

import { useModal } from "@/context/ModalContext";
import { NavItem } from "@/types/nav";
import { NavItemComponent } from "@/components/Navitem";
import LanguageDropdown from "@/components/LanguageDropdown";

export default function Navbar() {
    const { openModal } = useModal();

    const navItems: NavItem[] = [
        {
            name: "",
            icon: "home",
            href: "/",
        },
        {
            name: "Search",
            icon: "search",
            href: "/movies",
        },
        {
            name: "Sign In",
            icon: "user",
            action: () => openModal("signin"),
        },
        {
            name: "Create Account",
            icon: "register",
            action: () => openModal("register"),
        },
        {
            name: "",
            icon: "language",
            hover: LanguageDropdown,
        },
    ];

    return (<nav className="flex justify-between px-16 py-8">
        {navItems.map((item, index) => (
            <NavItemComponent key={index} item={item}/>
        ))}
    </nav>)
}
