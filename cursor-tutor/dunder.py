# 1. Try generating with command K. Ask for a pytorch script of a feedforward neural network
# 2. Then, select the outputted code and hit chat. Ask if there's a bug. Ask how to improve.
# 3. Try selecting some code and hitting edit. Ask the bot to add residual layers.
class Dunder:
    def __init__(self, value):
        self.value = value

    def __str__(self):
        return f"Dunder object with value: {self.value}"

    def __repr__(self):
        return f"Dunder({self.value})"

    def __add__(self, other):
        if isinstance(other, Dunder):
            return Dunder(self.value + other.value)
        return NotImplemented

    def __sub__(self, other):
        if isinstance(other, Dunder):
            return Dunder(self.value - other.value)
        return NotImplemented

    def __mul__(self, other):
        if isinstance(other, Dunder):
            return Dunder(self.value * other.value)
        return NotImplemented

    def __truediv__(self, other):
        if isinstance(other, Dunder):
            return Dunder(self.value / other.value)
        return NotImplemented

    def __eq__(self, other):
        if isinstance(other, Dunder):
            return self.value == other.value
        return NotImplemented

    def __ne__(self, other):
        if isinstance(other, Dunder):
            return self.value != other.value
        return NotImplemented

    def __lt__(self, other):
        if isinstance(other, Dunder):
            return self.value < other.value
        return NotImplemented

    def __le__(self, other):
        if isinstance(other, Dunder):
            return self.value <= other.value
        return NotImplemented

    def __gt__(self, other):
        if isinstance(other, Dunder):
            return self.value > other.value
        return NotImplemented

    def __ge__(self, other):
        if isinstance(other, Dunder):
            return self.value >= other.value
        return NotImplemented

    def __len__(self):
        return len(str(self.value))

    def __getitem__(self, index):
        return str(self.value)[index]

    def __setitem__(self, index, value):
        self.value = int(str(self.value)[:index] + str(value) + str(self.value)[index + 1:])

    def __delitem__(self, index):
        self.value = int(str(self.value)[:index] + str(self.value)[index + 1:])

    def __iter__(self):
        return iter(str(self.value))

    def __contains__(self, item):
        return str(item) in str(self.value)

    def __call__(self, new_value):
        self.value = new_value
        return self

if __name__ == "__main__":
    # Create a Dunder object and perform various operations
    dunder1 = Dunder(42)
    dunder2 = Dunder(13)

    print(dunder1 + dunder2)  # Add
    print(dunder1 - dunder2)  # Subtract
    print(dunder1 * dunder2)  # Multiply
    print(dunder1 / dunder2)  # Divide

    # Comparison
    print(dunder1 == dunder2)
    print(dunder1 != dunder2)
    print(dunder1 < dunder2)
    print(dunder1 <= dunder2)
    print(dunder1 > dunder2)
    print(dunder1 >= dunder2)

    # Length, indexing, and iteration
    print(len(dunder1))
    print(dunder1[1])
    dunder1[1] = 5
    print(dunder1)
    del dunder1[1]
    print(dunder1)

    for digit in dunder1:
        print(digit)

    # Check if a digit is in the Dunder object
    print(3 in dunder1)

    # Call the Dunder object to update its value
    print(dunder1(99))
