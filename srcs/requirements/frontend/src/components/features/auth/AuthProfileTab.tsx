import React from "react";
import useNotification from "@/contexts/NotificationContext";
import {useTranslations} from "next-intl";
import {postNewPassword} from "@/services/users.service";
import Form from "@/components/ui/Form";

export default function AuthProfileTab() {
    const {addNotification} = useNotification();
    const t = useTranslations("auth.changePassword");
    const tSuccess = useTranslations("notifications.success");

    const handlePasswordChange = () => {
        addNotification(tSuccess("passwordChanged"), "success");
    };

    return (<div className="max-w-9/10 sm:max-w-1/2 xl:max-w-2/6 w-full mx-auto flex flex-col items-start gap-4">
        <Form formType="auth" request={postNewPassword} handleRequest={handlePasswordChange} t={t}
              fields={["current-password", "new-password", "confirm-new-password"]} />
    </div>);
}
