import {iUser} from "@/types/user";
import React from "react";
import Image from "next/image";
import {useTranslations} from "next-intl";

export default function ProfilePicture({user, size = 0, color, className}: {user: iUser, size?: 0 | 1 | 2, color?: string, className?: string}) {
    const sizes = ["size-10", "size-18 sm:size-24", "size-38 sm:size-45"];
    const t = useTranslations("common");
    let children;

    if (user.profile_picture)
        children = <Image className="w-full h-full object-cover" height={200} width={200} src={user.profile_picture} alt={t("profilePictureAlt")} />;
    else {
        const initial = user.first_name[0] + user.last_name[0];

        if (color === undefined)
            color = user.color;

        if (size === 0)
            children = <h6>{initial}</h6>;
        else if (size === 1)
            children = <h4 className="text-3xl sm:text-5xl">{initial}</h4>;
        else
            children = <h1 className="text-6xl">{initial}</h1>;
    }

    return (<div className={`relative overflow-hidden rounded-full bg-${color} hover:bg-${color}-hover ${sizes[size]} flex items-center justify-center border shrink-0 ` + className}>
        <div className="custom-noise" />
        {children}
    </div>);
}
