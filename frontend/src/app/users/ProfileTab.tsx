import React, {useState} from "react";
import Input from "@/components/Input";
import Button from "@/components/Button";
import {tUser} from "@/types/user";
import ProfilePicture from "@/components/ProfilePicture";
import SmallButton from "@/components/SmallButton";


export default function ProfileTab({user, setUser}: {user: tUser, setUser: (tUser: tUser) => void}) {
    return (
        <div className="flex gap-30 max-w-2/3 w-full justify-center items-center mx-auto">
            <ProfileSection user={user} setUser={setUser} />
            <AvatarSection user={user} setUser={setUser} />
        </div>
    );
}

function ProfileSection({user, setUser}: {user: tUser, setUser: (tUser: tUser) => void}) {
    const [email, setEmail] = useState("");
    const [firstname, setFirstname] = useState("");
    const [lastname, setLastname] = useState("");
    const [username, setUsername] = useState("");

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>, key: string) => {
        const newUser = structuredClone(user);
        const newValue = e.target.value;

        if (key === "profile-email")
            setEmail(newValue);
        else if (key === "profile-firstname")
            setFirstname(newValue);
        else if (key === "profile-lastname")
            setLastname(newValue);
        else if (key === "profile-username")
            setUsername(newValue);
        setUser(newUser);
    }

    const saveChange = () => {
        const newUser = structuredClone(user);

        if (email)
            newUser.email = email;
        if (firstname)
            newUser.firstname = firstname;
        if (lastname)
            newUser.lastname = lastname;
        if (username)
            newUser.username = username;
        setEmail("");
        setFirstname("");
        setLastname("");
        setUsername("");
        setUser(newUser);
    }

    return (<div className="flex flex-col gap-4 items-start">
        <Input id="profile-email" type="email" placeholder="Email" value={email} onChange={handleChange}></Input>

        <div className="flex gap-2 w-full">
            <Input id="profile-firstname" type="firstname" placeholder="Firstname" value={firstname} onChange={handleChange}></Input>
            <Input id="profile-lastname" type="lastname" placeholder="Lastname" value={lastname} onChange={handleChange}></Input>
        </div>

        <Input id="profile-username" type="username" placeholder="Username" value={username} onChange={handleChange} className={"max-w-3/5"}></Input>

        <Button className="h-8" onClick={saveChange}>Save Changes</Button>
    </div>);
}


function AvatarSection({user, setUser}: {user: tUser, setUser: (tUser: tUser) => void}) {
    const colors = ["yellow", "pink", "green", "purple", "blue", "red"];

    const handleNewPP = (newPP: string | null) => {
        const newUser = structuredClone(user);

        newUser.profile_picture = newPP;
        setUser(newUser);
    }

    const handleSwitchColors = (newColor: string) => {
        const newUser = structuredClone(user);

        newUser.color = newColor;
        setUser(newUser);
    }

    const uploadNewPP = () => {
        handleNewPP("/images/profile_pictures.jpeg");
    }

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
