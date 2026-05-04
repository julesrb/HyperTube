"use client";

import React from "react";
import {tNotification, useNotification} from "@/context/NotificationContext";
import {CheckIcon} from "@/components/Icons";
import {CloseButton} from "./Buttons";

export const NotificationList = () => {
    const { notifications, removeNotification } = useNotification();

    return (<div className="fixed top-5 right-5 z-60 flex flex-col gap-4 max-w-9/10">
            {notifications.map((notification) => (
                <NotificationItem
                    key={notification.id}
                    notification={notification}
                    onClose={removeNotification}
                />
            ))}
        </div>);
};

const NotificationItem = ({notification, onClose}: { notification: tNotification; onClose: (id: string) => void;}) => {
    const bgColors = {success: "green", warning: "yellow", error: "red", info: "purple"};

    return (<div className={`flex justify-between custom-shadow-s border bg-${bgColors[notification.type]}`}>

            <div className="flex items-center sm:min-w-90 max-w-90 border-r p-2 sm:p-4 leading-tight sm:leading-normal">
                {notification.type === "success" && <CheckIcon className="shrink-0 mr-2 sm:mr-4"/>}
                <p>{notification.message}</p>
            </div>

            <CloseButton className="p-2 sm:p-4" onClick={() => onClose(notification.id)}/>
        </div>
    );
};
