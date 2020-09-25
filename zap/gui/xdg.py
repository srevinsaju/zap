from gi.repository import Gtk, GLib
import os
import threading
import queue
import gi
import requests

gi.require_version('Gtk', '3.0')


# https://stackoverflow.com/q/55868685


class Downloader(threading.Thread):
    def __init__(self, root, queue, url, output_directory, filename=None):
        threading.Thread.__init__(self)
        self.root = root
        self._queue = queue
        self.filename = filename
        self.url = url
        self.local_filename = None
        self.output_directory = output_directory

    def run(self):
        if not self.filename:
            local_filename = \
                os.path.join(self.output_directory, self.url.split('/')[-1])
        else:
            local_filename = os.path.join(self.output_directory, self.filename)
        r = requests.get(self.url, stream=True)
        file_size = int(r.headers['Content-Length'])
        chunk = 1
        chunk_size = 1024
        num_bars = int(file_size / chunk_size)

        with open(local_filename, 'wb') as fp:
            for chunk in r.iter_content(chunk_size=chunk_size):
                # for chunk in range(num_bars):
                fp.write(chunk)
                self._queue.put((1 / num_bars) * 100)
        self.root.local_filename = local_filename


class ZapXDGDownloader(Gtk.Window):
    def __init__(self, **kwargs):
        Gtk.Window.__init__(self)
        self.set_size_request(300, 150)
        self.set_border_width(10)

        vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=6)
        self.add(vbox)

        self.label = Gtk.Label("Downloading AppImage 🗲")
        vbox.pack_start(self.label, True, True, 0)

        # max and current number of tasks
        self._max = 100
        self._curr = 0
        self.local_filename = None
        # queue to share data between threads
        self._queue = queue.Queue()

        # gui: progressbar
        self._bar = Gtk.ProgressBar(show_text=True)
        vbox.add(self._bar)
        self.connect('destroy', Gtk.main_quit)

        # install timer event to check the queue for new data from the thread
        GLib.timeout_add(interval=250, function=self._on_timer)
        # start the thread
        self._thread = Downloader(self, self._queue, **kwargs)
        self._thread.start()

    def _on_timer(self):
        # if the thread is dead and no more data available...
        if not self._thread.is_alive() and self._queue.empty():
            # ...end the timer
            Gtk.main_quit()
            return False

        # if data available
        while not self._queue.empty():
            # read data from the thread
            self._curr += self._queue.get()
            # update the progressbar
            self._bar.set_fraction(self._curr / self._max)

        # keep the timer alive
        return True


def gtk_zap_downloader(url, output_directory, filename=None, **kwargs):
    win = ZapXDGDownloader(url=url, filename=filename,
                           output_directory=output_directory)
    win.connect("destroy", Gtk.main_quit)
    win.show_all()
    Gtk.main()
    if win.local_filename:
        return win.local_filename
    else:
        raise RuntimeError("localFilename returned falsey value.")


if __name__ == '__main__':
    url = input("Enter URL to download: ")
    filename = input("Enter filename: ")
    od = input("Enter output directory: ")
    gtk_zap_downloader(url, filename, od)
