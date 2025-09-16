/*function Profile() {
  return <img src="https://i.imgur.com/MK3eW3As.jpg" alt="Katherine Johnson" />;
}*/
import Profile from "./Profile";
import TodoList from "./Todo";

import List from "./List";
import { people } from "./data.js";
import VideoPlayerComponent from "./VideoPlayerComponent.js";
import MyToolBar from "./Toolbar.js";
import MyGallery from "./MyGallery.js";
import ChatForm from "./ChatForm.js";
import Counter from "./Counter.js";
import PersonForm from "./PersonForm.js";
import BucketList from "./BucketList.js";
import AutoEnableForm from "./AutoEnableForm.js";
import FocusForm from "./FocusForm.js";

export default function Gallery() {
  return (
    <>
      <section>
        <h1>Amazing scientists</h1>
        <Profile />
        <Profile />
        <Profile />
      </section>
      <TodoList />

      <List data={people} />
      <MyGallery />
      <MyToolBar />
      <ChatForm />
      <Counter />
      <PersonForm />
      <BucketList />
      <AutoEnableForm />
      <FocusForm />
      <p></p>
      <div>
        <VideoPlayerComponent />
      </div>
    </>
  );
}
