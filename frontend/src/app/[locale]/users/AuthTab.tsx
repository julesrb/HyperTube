import React, {useState} from "react";
import Input from "@/components/Input";
import {useNotification} from "@/context/NotificationContext";
import {Button} from "@/components/Buttons";
import {useTranslations} from "next-intl";

export default function AuthTab() {
    const { addNotification } = useNotification();
    const [oldPassword, setOldPassword] = useState("");
    const [newPassword, setNewPassword] = useState("");
    const [confirmNewpassword, setConfirmNewPassword] = useState("");
    const t = useTranslations("auth.changePassword");
    const tError = useTranslations("notifications.error");
    const tSuccess = useTranslations("notifications.success");

    const saveChange = () => {
        if (oldPassword === "" || newPassword === "" || confirmNewpassword === "")
            addNotification(tError("requiredFields"), "error");
        else if (false)
            addNotification(tError("passwordIncorrect"), "error");
        else if (newPassword !== confirmNewpassword)
            addNotification(tError("passwordMismatch"), "error");
        else if (oldPassword !== newPassword)
            addNotification(tError("passwordSameAsOld"), "error");
        else {
            setOldPassword("");
            setNewPassword("");
            setConfirmNewPassword("");
            addNotification(tSuccess("passwordChanged"), "success");
        }
    }

    return (<div className="max-w-9/10 sm:max-w-1/2 xl:max-w-2/6 w-full mx-auto flex flex-col items-start gap-4">
        <Input id="old-password" type="password" value={oldPassword} onChange={setOldPassword} placeholder={t("current")}></Input>
        <Input id="new-password" type="password" value={newPassword} onChange={setNewPassword} placeholder={t("new")}></Input>
        <Input id="confirm-new-password" type="password" value={confirmNewpassword} onChange={setConfirmNewPassword} placeholder={t("confirm")}></Input>
        <Button onClick={saveChange}>{t("submit")}</Button>
    </div>);
}
