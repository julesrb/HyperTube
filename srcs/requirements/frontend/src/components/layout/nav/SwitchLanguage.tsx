import React, {useState} from "react";
import {useRouter, usePathname} from "@/i18n/navigation";
import {tLocale} from "@/i18n/request";
import {useSearchParams} from "next/navigation";
import LanguageDropdown from "@/components/LanguageDropdown";
import {useLocale} from "next-intl";

export default function SwitchLanguage(Icon: ({selected}: {selected: boolean}) => React.JSX.Element) {
    const [isOpen, setIsOpen] = useState(false);
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();
    const locale = useLocale() as tLocale;

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
        {isOpen && <LanguageDropdown handleSwitchLanguage={handleSwitchLanguage} selected={locale} className="top-10 right-1/30" />}
    </div>);
}
