"use client";

import React, {useEffect} from "react";
import {useTranslations} from "next-intl";
import {useRouter} from "@/i18n/navigation";
import useAuth from "@/contexts/AuthContext";
import useNotification from "@/contexts/NotificationContext";
import SmallText from "@/components/ui/SmallText";

export default function OAuthCallbackPage() {
    const t = useTranslations("auth.oauth");
    const {login} = useAuth();
    const router = useRouter();
    const {addNotification} = useNotification();
    const tError = useTranslations("notifications.error");

    useEffect(() => {
        const hash = window.location.hash;
        const params = new URLSearchParams(hash.replace("#", ""));
        const token = params.get("access_token");
        const refresh = params.get("refresh_token");
        const userEncoded = params.get("user");
        const redirectParam = params.get("redirect");
        let redirect = "/";

        try {
            if (!userEncoded || !token || !refresh)
                addNotification(tError("invalidQueryParameter"), "error");
            else {
                const user = JSON.parse(decodeURIComponent(decodeURIComponent(userEncoded)));
                login(user, token, refresh);
                if (redirectParam)
                    redirect = decodeURIComponent(redirectParam);
            }
        } catch {
            addNotification(tError("invalidQueryParameter"), "error");
        }
        router.push(redirect);
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [tError]);

    return (<SmallText>{t("loadingAuth")}</SmallText>);
}
