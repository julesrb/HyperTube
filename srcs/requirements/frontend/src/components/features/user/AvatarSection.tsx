import {useTranslations} from "next-intl";
import {iUser} from "@/types/user";
import ProfilePicture from "@/components/features/user/ProfilePicture";
import TextButton from "@/components/ui/Button/TextButton";
import useApiMutation from "@/hooks/useApiMutation";
import {patchUser} from "@/services/users.service";

export default function AvatarSection({user, updateUser}: {user: iUser, updateUser?: (patch: Partial<iUser>) => void}) {
    const colors = ["yellow", "pink", "green", "purple", "blue", "red"];
    const t = useTranslations("profile");
    const {execute} = useApiMutation();

    const handleNewAvatar = (newData: string[]) => {
        const makePostRequest = async () => {
            return await execute((locale) => patchUser(locale, newData, user.id));
        };

        makePostRequest().then((data) => {
            if (data) {
                const newPartialUser: Record<string, string> = {};
                newPartialUser[newData[0]] = newData[1];
                if (updateUser)
                    updateUser(newPartialUser);
            }
        })
    }

    return (<div className="flex flex-col gap-2 items-center justify-center">
        <ProfilePicture user={user} size={2} className="mb-2" />
        <TextButton
            className={user.profile_picture ? "text-red custom-underline-red" : "hover:no-underline"}
            onClick={() => handleNewAvatar(["profile_picture", ""])}>{t("remove")}</TextButton>

        {!user.profile_picture && (
            <div className="grid grid-cols-3 gap-2 mt-4">
                {colors.map((color, index) => (
                    <button key={index} onClick={() => handleNewAvatar(["color", color])}>
                        <ProfilePicture user={user} color={color} className={user.color === color ? "border-3" : ""}/>
                    </button>
                ))}
            </div>)}
    </div>);
}
