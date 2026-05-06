import React, {useState} from "react";
import {useRouter, usePathname} from '@/i18n/navigation';
import {routing, tLocale} from "@/i18n/request";

export default function LanguageDropdown(Icon: ({selected}: {selected: boolean}) => React.JSX.Element) {
    const [isOpen, setIsOpen] = useState(false);
    const languages: Record<tLocale, string> = {en: "English", fr: "Français", de: "Deutsch"};
    const [selectedLanguage, setSelectedLanguage] = useState<Record<tLocale, boolean>>({en: false, fr: false, de: false});
    const router = useRouter();
    const pathname = usePathname();

    const handleSwitchLanguage = (key: tLocale) => {
        const newLanguage = {en: false, fr: false, de: false};
        newLanguage[key] = true;
        setSelectedLanguage(newLanguage);
        router.replace(pathname, {locale: key});
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
                    <button key={code} className={"custom-underline text-xl " + (selectedLanguage[code] ? "font-base font-light" : "font-hairline")} onClick={() => handleSwitchLanguage(code)}>{languages[code]}</button>
                ))}
            </div>
        </div>}
    </div>);
}
