import React, {useState} from "react";
import {useRouter, usePathname} from '@/i18n/navigation';
import {routing, tLocale} from "@/i18n/request";
import {useLocale} from "next-intl";
import {useSearchParams} from "next/navigation";

export default function LanguageDropdown(Icon: ({selected}: {selected: boolean}) => React.JSX.Element) {
    const [isOpen, setIsOpen] = useState(false);
    const languages: Record<tLocale, string> = {en: "English", fr: "Français", de: "Deutsch"};
    const locale = useLocale() as tLocale;
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();

    const handleSwitchLanguage = (key: tLocale) => {
        const query = searchParams.toString();
        const href = query.length > 0 ? `${pathname}?${query}` : pathname;
        router.replace(href, {locale: key});
        setIsOpen(false);
    }

    return (<div
        className="flex items-center hover:cursor-pointer"
        onMouseEnter={() => (setIsOpen(true))}
        onMouseLeave={() => (setIsOpen(false))}>

        <Icon selected={isOpen} />
        {isOpen && <div className="absolute top-10 right-1/30 z-50 p-8">
            <div className="flex flex-col gap-1 items-start bg-white py-4 px-5 custom-shadow-s border">
                {routing.locales.map((code) => (
                    <button key={code} className={"custom-underline text-xl " + (locale === code ? "font-base font-light" : "font-hairline")} onClick={() => handleSwitchLanguage(code)}>{languages[code]}</button>
                ))}
            </div>
        </div>}
    </div>);
}
