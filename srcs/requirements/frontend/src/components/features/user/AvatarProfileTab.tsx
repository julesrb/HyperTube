import AvatarSection from "@/components/features/user/AvatarSection";
import {iUser} from "@/types/user";

export default function AvatarProfileTab({user, updateUser}: {user: iUser, updateUser?: (patch: Partial<iUser>) => void}) {
    return (<div className="max-w-9/10 sm:max-w-1/2 xl:max-w-2/6 w-full mx-auto">
        <AvatarSection user={user} updateUser={updateUser} />
    </div>);
}
