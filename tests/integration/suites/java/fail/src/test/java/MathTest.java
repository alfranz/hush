import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertEquals;

class MathTest {

    @Test
    void addition() {
        assertEquals(2, 1 + 1);
    }

    @Test
    void badMath() {
        assertEquals(99, 2 + 2);
    }
}
