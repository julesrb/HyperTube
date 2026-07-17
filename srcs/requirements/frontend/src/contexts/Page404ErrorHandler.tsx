"use client";

import {useSearchParams} from "next/navigation";
import useNotification from "@/contexts/NotificationContext";
import {useEffect} from "react";
import {useTranslations} from "next-intl";
import {usePathname, useRouter} from "@/i18n/navigation";

export default function Page404ErrorHandler() {
    const params = useSearchParams();
    const {addNotification} = useNotification();
    const tError = useTranslations("notifications.error");
    const pathname = usePathname();
    const router = useRouter();

    useEffect(() => {
        const errorPathname = params.get("error");
        if (errorPathname) {
            addNotification(tError("notFound", {path: errorPathname}), "error");
            router.replace(pathname);
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [params]);

    return null;
}
