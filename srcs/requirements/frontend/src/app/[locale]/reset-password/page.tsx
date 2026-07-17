"use client";

import {useTranslations} from "next-intl";
import {useRouter} from "@/i18n/navigation";
import {useSearchParams} from "next/navigation";
import {useEffect} from "react";
import useNotification from "@/contexts/NotificationContext";
import useModal from "@/contexts/ModalContext";
import useAuth from "@/contexts/AuthContext";

export default function Page() {
    const searchParams = useSearchParams();
    const token = searchParams.get("token");
    const {addNotification} = useNotification();
    const router = useRouter();
    const {openModal} = useModal();
    const {user, loading} = useAuth();
    const tError = useTranslations("notifications.error");

    useEffect(() => {
        if (!loading) {
            if (!user) {
                if (token)
                    openModal({type: "set-new-password", token: token});
                else
                    addNotification(tError("missingToken"), "error");
            }
            setTimeout(() => router.replace("/"), 0);
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [loading]);
}
