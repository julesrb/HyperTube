import {cookies} from 'next/headers';
import {getRequestConfig} from 'next-intl/server';
import {defineRouting} from 'next-intl/routing';
import {hasLocale} from "next-intl";

export const routing = defineRouting({
    locales: ['en', 'fr', 'de'],

    defaultLocale: 'en'
});

export default getRequestConfig(async ({requestLocale}) => {
    const store = await cookies();
    const requested = await requestLocale;
    let locale = store.get('locale')?.value;
    if (!hasLocale(routing.locales, locale)) {
        locale = hasLocale(routing.locales, requested)
            ? requested
            : routing.defaultLocale;
    }

    return {
        locale,
        messages: (await import(`../../messages/${locale}.json`)).default
    };
});
