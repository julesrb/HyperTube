"use client";

import React, {JSX, useState} from "react";
import ProfilePicture from "@/components/features/user/ProfilePicture";
import {iUser} from "@/types/user";
import {useSearchParams} from "next/navigation";
import {useLocale, useTranslations} from "next-intl";
import {usePathname, useRouter} from "@/i18n/navigation";

export type tTab = {
    name: string
    comp: ({user, updateUser}: {user: iUser, updateUser?: (patch: Partial<iUser>) => void}) => JSX.Element}[];

export default function UserProfile({user, tabs, updateUserAction}: {user: iUser, tabs: tTab, updateUserAction?: (patch: Partial<iUser>) => void}) {
    const searchParams = useSearchParams();
    const router = useRouter();
    const pathname = usePathname();
    const tabParam = searchParams.get("tab");
    let initialTab: number = 0;
    if (tabParam) {
        const idx = tabs.findIndex(a => a.name === tabParam)
        if (idx >= 0)
            initialTab = idx;
    }
    const [activeTab, setActiveTab] = useState<number>(initialTab);
    const t = useTranslations("profile.tabs");
    const tProfile = useTranslations("profile");
    const locale = useLocale();
    const ActiveTab = tabs[activeTab].comp;
    const showLastName = user.first_name.length > 1 && user.last_name.length > 1;
    const name = showLastName ? `${user.first_name} ${user.last_name[0]}.` : user.username;

    const date = new Date(user.created_at);
    const memberSince = new Intl.DateTimeFormat(locale, {day: "2-digit", month: "2-digit", year: "numeric"}).format(date).replace(/[\/-]/g, ".");

    const switchTab = (tabIdx: number) => {
        if (activeTab !== tabIdx) {
            const params = new URLSearchParams(searchParams.toString());
            params.set("tab", tabs[tabIdx].name);
            router.push(`${pathname}?${params.toString()}`);
            setActiveTab(tabIdx);
        }
    }

    return (<div className="flex flex-col gap-6 sm:gap-12 xl:gap-17 px-2 md:px-4">
        <div className="flex items-center gap-4 justify-center">
            <ProfilePicture user={user} size={1}/>
            <div className="flex flex-col items-start">
                <h2>{name}</h2>
                <p className="uppercase">{tProfile("memberSince", {date: memberSince}) + (showLastName ? ` · #${user.username}` : "")}</p>
            </div>
        </div>
        <div className="flex h-12 sm:h-16">
            <div className="border-b border-r w-12" />
            {tabs.map((tab, index) => (<button
                key={index}
                className={"custom-condensed text-2xl sm:text-3xl md:text-4xl tracking-wide sm:tracking-normal border-t border-r px-3 sm:px-12 xl:px-16 border-b text-nowrap" + (activeTab === index ? " border-b-white" : "")}
                onClick={() => switchTab(index)}>{t(tab.name)}</button>))}
            <div className="border-b w-full" />
        </div>
        <ActiveTab user={user} updateUser={updateUserAction}/>
        <div />
    </div>);
}
