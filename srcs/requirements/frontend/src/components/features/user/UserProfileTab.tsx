import AvatarSection from "@/components/features/user/AvatarSection";
import ProfileSection from "@/components/features/user/ProfileSection";
import {iUser} from "@/types/user";

export default function UserProfileTab({user, updateUser}: {user: iUser, updateUser?: (patch: Partial<iUser>) => void}) {
    return (<div className="flex flex-col sm:flex-row gap-14 sm:gap-20 xl:gap-30 max-w-9/10 xl:max-w-2/3 w-full justify-center items-center mx-auto">
        <ProfileSection user={user} updateUser={updateUser} />
        <AvatarSection user={user} updateUser={updateUser} />
    </div>);
}
