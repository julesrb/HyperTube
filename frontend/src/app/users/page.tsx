import Input from "@/components/Input";
import React from "react";
import Button from "@/components/Button";
import DefaultUserIcon from "@/components/icon/DefaultUserIcon";
import FilmSmallCard from "@/components/MovieSmallCard";
import {movies} from "@/types/movie";

export default function Page() {

    return (<div>
            <div className="container mx-auto">
                <h3>Profile</h3>
                <div className="grid grid-cols-2">
                    <div className="flex flex-col gap-4 items-start max-w-2/3 mx-auto">
                        <Input type="email" placeholder="Email"></Input>

                        <div className="flex gap-2 w-full">
                            <Input type="firstname" placeholder="Firstname"></Input>
                            <Input type="lastname" placeholder="Lastname"></Input>
                        </div>

                        <Input type="username" placeholder="Username" className={"max-w-2/3"}></Input>

                        <Button className="h-8">Save Changes</Button>
                    </div>

                    <div className="flex flex-col gap-2 items-center justify-center">
                        <DefaultUserIcon className="mb-6"/>
                        <Button className="h-8">Select New avatar</Button>

                        <button
                            className={(true ? "text-red  custom-underline-red" : "text-gray") + " text-sm font-light"}>Remove
                        </button>
                    </div>
                </div>
            </div>
            <div className="container mx-auto">
                <h3>Password</h3>

                <Input type="password" placeholder="Current password"></Input>
                <Input type="password" placeholder="New password"></Input>
                <Input type="password" placeholder="Confirm new password"></Input>

                <Button className="h-8">Change</Button>
            </div>
            <div className="container mx-auto">
                <h3>Films Watched</h3>
                <div className="grid grid-cols-4">
                    {movies.map((movie, index) => (
                        <FilmSmallCard key={index} movie={movie}/>))}
                </div>
            </div>
        </div>);
}
