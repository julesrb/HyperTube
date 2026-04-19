import React, {useState} from "react";
import {Button} from "@/components/Button";
import Input from "@/components/Input";
import {useNotification} from "@/context/NotificationContext";
import {errorMessages, successMessages} from "@/types/message";

export default function AuthTab() {
    const { addNotification } = useNotification();
    const [oldPassword, setOldPassword] = useState("");
    const [newPassword, setNewPassword] = useState("");
    const [confirmNewpassword, setConfirmNewPassword] = useState("");

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>, id: string) => {
        const newValue = e.target.value.trim();

        if (id === "old-password")
            setOldPassword(newValue);
        else if (id === "new-password")
            setNewPassword(newValue);
        else
            setConfirmNewPassword(newValue);
    }

    const saveChange = () => {
        if (oldPassword === "" || newPassword === "" || confirmNewpassword === "")
            addNotification(errorMessages.requiredFields, "error");
        else if (false)
            addNotification(errorMessages.passwordIncorrect, "error");
        else if (newPassword !== confirmNewpassword)
            addNotification(errorMessages.passwordMismatch, "error");
        else if (oldPassword !== newPassword)
            addNotification(errorMessages.passwordSameAsOld, "error");
        else {
            setOldPassword("");
            setNewPassword("");
            setConfirmNewPassword("");
            addNotification(successMessages.passwordChanged, "success");
        }
    }

    return (<div className="max-w-2/6 mx-auto flex flex-col items-start gap-4">

        <Input id="old-password" type="password" value={oldPassword} onChange={handleChange} placeholder="Current password"></Input>
        <Input id="new-password" type="password" value={newPassword} onChange={handleChange} placeholder="New password"></Input>
        <Input id="confirm-new-password" type="password" value={confirmNewpassword} onChange={handleChange} placeholder="Confirm new password"></Input>

        <Button onClick={saveChange}>Change</Button>
    </div>);
}
