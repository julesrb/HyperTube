import React from "react";
import {iUser, iUserToken} from "@/types/user";
import {useRouter} from "@/i18n/navigation";
import {useTranslations} from "next-intl";
import useModal, {ModalState} from "@/contexts/ModalContext";
import useNotification from "@/contexts/NotificationContext";
import useAuth from "@/contexts/AuthContext";
import {tResponse} from "@/types/api";
import ModalLayout from "@/components/layout/ModalLayout";
import Form from "@/components/ui/Form";
import {postLogin, postRegister} from "@/services/auth.service";
import TextButton from "@/components/ui/Button/TextButton";

type AuthModalType = "signin" | "register";

export default function AuthModalLayout({type, t, handleForgotPassword, activeModal}: {type: AuthModalType, t: (key: string) => string, handleForgotPassword?: () => void, activeModal: ModalState}) {
    const {openModal, closeModal} = useModal();
    const isReg = type === "register";
    const otherType: AuthModalType = isReg ? "signin" : "register";
    const {addNotification} = useNotification();
    const router = useRouter();
    const {login, callbackUrl, setCallbackUrl} = useAuth();
    const tSuccess = useTranslations("notifications.success");

    const handleLoginRegister = (data: tResponse<iUserToken | iUser>) => {
        if ("access_token" in data.data) {
            login(data.data.user, data.data.access_token, data.data.refresh_token);
            closeModal();
            if (callbackUrl) {
                router.push(callbackUrl);
                setCallbackUrl(undefined);
            }
            addNotification(tSuccess(isReg ? "accountCreatedSuccess" : "login"), "success");
            if (activeModal.reload)
                activeModal.reload();
        }
    };

    return (<ModalLayout onCloseAction={activeModal.noClose === undefined ? () => {
        if (activeModal.noClose === undefined) {
            setCallbackUrl(undefined);
            closeModal();
        }
    } : undefined} title={t("title" + (callbackUrl ? "LoginRequired" : ""))}>
        <Form formType={type} request={isReg ? postRegister : postLogin} handleRequest={handleLoginRegister} t={t} handleForgotPassword={handleForgotPassword}
              fields={isReg ? ["email", "first_name", "last_name", "username", "password"] : ["login", "password"]} />
        <div className="flex gap-2 mt-2">
            <span className="text-sm">{t(isReg ? "haveAccount" : "noAccount")}</span>
            <TextButton onClick={() => {closeModal(); openModal({type: otherType, noClose: activeModal.noClose});}}>{t(otherType)}</TextButton>
        </div>
    </ModalLayout>);
}
