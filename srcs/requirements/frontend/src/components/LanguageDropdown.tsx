import React from "react";
import {routing, tLocale} from "@/i18n/request";

export default function LanguageDropdown({handleSwitchLanguage, selected, className, strikethrough=false}: {handleSwitchLanguage: (l: tLocale) => void, selected: tLocale | undefined, className: string, strikethrough?: boolean}) {
    const languages: Record<tLocale, string> = {en: "English", fr: "Français", de: "Deutsch"};

    return (<div className={"absolute z-50 py-6 px-8 " + className}>
        <div className="flex flex-col gap-1 items-start bg-white py-4 px-5 custom-shadow-m border border-black">
            {routing.locales.map((code) => (
                <button key={code} className={"text-black text-xl " + (selected === code ? "font-base font-light" : "font-hairline") + ((selected === code && strikethrough) ? " hover:line-through" : " custom-underline")} onClick={() => handleSwitchLanguage(code)}>{languages[code]}</button>
            ))}
        </div>
    </div>);
}
