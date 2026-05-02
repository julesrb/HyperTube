import React, {useState} from "react";
import Input from "@/components/Input";
import {tUser} from "@/types/user";
import ProfilePicture from "@/components/ProfilePicture";
import {useNotification} from "@/context/NotificationContext";
import {successMessages} from "@/types/message";
import {Button, SmallButton} from "@/components/Buttons";


export default function ProfileTab({user, updateUser}: {user: tUser, updateUser: (patch: Partial<tUser>) => void}) {
    return (<div className="mb-20 flex flex-col sm:flex-row gap-14 sm:gap-20 xl:gap-30 max-w-9/10 xl:max-w-2/3 w-full justify-center items-center mx-auto">
        <ProfileSection user={user} updateUser={updateUser} />
        <AvatarSection user={user} updateUser={updateUser} />
    </div>);
}

function ProfileSection({user, updateUser}: {user: tUser, updateUser: (patch: Partial<tUser>) => void}) {
    const { addNotification } = useNotification();
    const [email, setEmail] = useState("");
    const [firstname, setFirstname] = useState("");
    const [lastname, setLastname] = useState("");
    const [username, setUsername] = useState("");

    const saveChange = () => {
        const newUser = structuredClone(user);
        let isInfoChanged = false;

        if (email && email != user.email) {
            newUser.email = email;
            addNotification(successMessages.emailChanged, "warning");
            setEmail("");
        }
        if (firstname && firstname != user.firstname) {
            isInfoChanged = true;
            newUser.firstname = firstname;
            setFirstname("");
        }
        if (lastname && lastname != user.lastname) {
            isInfoChanged = true;
            newUser.lastname = lastname;
            setLastname("");
        }
        if (username && username != user.username) {
            isInfoChanged = true;
            newUser.username = username;
            setUsername("");
        }
        if (isInfoChanged) {
            updateUser(newUser);
            addNotification(successMessages.infoChanged, "success");
        }
    }

    return (<div className="flex flex-col gap-4 items-start">
        <Input id="profile-email" type="email" placeholder="Email" value={email} onChange={(newValue) => setEmail(newValue)}></Input>

        <div className="flex gap-2 w-full">
            <Input id="profile-firstname" type="firstname" placeholder="Firstname" value={firstname} onChange={(newValue) => setFirstname(newValue)}></Input>
            <Input id="profile-lastname" type="lastname" placeholder="Lastname" value={lastname} onChange={(newValue) => setLastname(newValue)}></Input>
        </div>

        <Input id="profile-username" type="username" placeholder="Username" value={username} onChange={(newValue) => setUsername(newValue)} className={"max-w-3/5"}></Input>

        <Button className="h-8" onClick={saveChange}>Save Changes</Button>
    </div>);
}


function AvatarSection({user, updateUser}: {user: tUser, updateUser: (patch: Partial<tUser>) => void}) {
    const colors = ["yellow", "pink", "green", "purple", "blue", "red"];

    const handleNewPP = (newPP: string | null) => {updateUser({profile_picture: newPP});}
    const handleSwitchColors = (newColor: string) => {updateUser({color: newColor});}
    const uploadNewPP = () => {handleNewPP("/images/profile_pictures.jpeg");}

    return (<div className="flex flex-col gap-2 items-center justify-center">
        <ProfilePicture user={user} size={2} className="mb-6" onClick={uploadNewPP}/>
        <Button onClick={uploadNewPP}>Select New avatar</Button>

        <SmallButton
            className={user.profile_picture ? "text-red  custom-underline-red" : "custom-no-underline"}
            onClick={() => handleNewPP(null)}>Remove</SmallButton>

        { !user.profile_picture && (
            <div className="grid grid-cols-3 gap-2 mt-4">
                {colors.map((color, index) => (
                    <ProfilePicture
                        key={index}
                        user={user}
                        color={color}
                        className={user.color === color ? "border-3" : ""}
                        onClick={() => handleSwitchColors(color)}
                    />))}
            </div>)}
    </div>);
}
