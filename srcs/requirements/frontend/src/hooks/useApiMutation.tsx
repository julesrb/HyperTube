"use client";

import {useLocale, useTranslations} from "next-intl";
import {tLocale} from "@/i18n/request";
import {ApiError} from "@/services/ApiError";
import useNotification from "@/contexts/NotificationContext";
import useModal from "@/contexts/ModalContext";
import {fieldType} from "@/components/ui/Form";
import useAuth from "@/contexts/AuthContext";
import {usePathname} from "@/i18n/navigation";

export default function useApiMutation(setErrorsAction?: (errors: Record<string, string>) => void, setFocusedIndex?: (idx: number) => void, formType?: string, fields?: fieldType[]) {
    const {setCallbackUrl, logout} = useAuth();
    const {openModal, closeModal} = useModal();
    const locale = useLocale() as tLocale;
    const {addNotification} = useNotification();
    const tError = useTranslations("notifications.error");
    const pathname = usePathname();

    async function execute<T>(callback: (locale: string) => Promise<T>): Promise<T | null> {
        try {
            return await callback(locale);
        } catch (error) {
            if (error instanceof ApiError) {
                if (error.data.error.fields && setErrorsAction && formType && fields) {
                    const newErrors: Record<string, string> = {};
                    let setNewFocus = false;

                    fields.forEach((field, idx)=> {
                        if (error.data.error.fields && error.data.error.fields[field]) {
                            newErrors[field + "-" + formType] = error.data.error.fields[field].message;
                            if (!setNewFocus && setFocusedIndex) {
                                setNewFocus = true;
                                setFocusedIndex(idx);
                            }
                        }
                    });
                    setErrorsAction(newErrors);
                    return null;
                } else if (error.data.error.code === "INVALID_RESET_TOKEN") {
                    closeModal();
                    addNotification(tError("invalidToken"), "error");
                    return null;
                } else if (error.status === 401 && setErrorsAction && setFocusedIndex && formType === "signin") {
                    setErrorsAction({"login-signin": error.message});
                    setFocusedIndex(0);
                    return null;
                } else if (error.status === 401) {
                    setCallbackUrl(pathname);
                    openModal({type: "signin"});
                    logout();
                    return null;
                } else
                    addNotification(error.notificationMsg, "error");
            } else
                addNotification(tError("network"), "error");
            return null;
        }
    }
    return {execute};
}
