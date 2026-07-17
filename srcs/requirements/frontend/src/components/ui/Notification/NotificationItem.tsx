import {CheckIcon} from "@/components/Icons";
import CloseButton from "@/components/ui/Button/CloseButton";
import {tNotification} from "@/contexts/NotificationContext";

export default function NotificationItem({notification, onClose}: {notification: tNotification; onClose: (id: string) => void})  {
    const bgColors = {success: "green", warning: "yellow", error: "red", info: "purple"};

    return (<div className={`flex justify-between custom-shadow-m border bg-${bgColors[notification.type]}`}>

            <div className="flex items-center sm:min-w-90 max-w-90 border-r p-2 sm:p-4 leading-tight sm:leading-normal">
                {notification.type === "success" && <CheckIcon className="shrink-0 mr-2 sm:mr-4"/>}
                <p>{notification.message}</p>
            </div>

            <CloseButton className="p-2 sm:p-4" onClickAction={() => onClose(notification.id)}/>
        </div>
    );
}
